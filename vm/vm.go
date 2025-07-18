package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const InitialStackSize = 256 // Start with smaller stack, grow as needed
const MaxStackSize = 2048    // Maximum stack size (original StackSize)
const GlobalsSize = 65536    // Keep globals size for now
const MaxFrames = 1024

// Use shared singleton instances from object package to reduce memory allocation

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack    []object.Object
	stackCap int // Current stack capacity
	sp       int
	globals  []object.Object

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack:      make([]object.Object, InitialStackSize),
		stackCap:   InitialStackSize,
		sp:         0,
		globals:    make([]object.Object, GlobalsSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		op = code.Opcode(ins[ip])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(object.TRUE)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(object.FALSE)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpLessThan, code.OpGreaterThanEqual, code.OpLessThanEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpLogicalAnd:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			left := vm.pop()
			if !isTruthy(left) {
				vm.currentFrame().ip = pos - 1
				err := vm.push(left)
				if err != nil {
					return err
				}
			}
		case code.OpLogicalOr:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			left := vm.pop()
			if isTruthy(left) {
				vm.currentFrame().ip = pos - 1
				err := vm.push(left)
				if err != nil {
					return err
				}
			}
		case code.OpNull:
			err := vm.push(object.NULL)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(object.NULL)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.FreeVariables[freeIndex])
			if err != nil {
				return err
			}
		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		}
	}
	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= vm.stackCap {
		if err := vm.growStack(); err != nil {
			return err
		}
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

// growStack doubles the stack size up to MaxStackSize
func (vm *VM) growStack() error {
	newCap := vm.stackCap * 2
	if newCap > MaxStackSize {
		return fmt.Errorf("stack overflow: maximum stack size (%d) exceeded", MaxStackSize)
	}

	newStack := make([]object.Object, newCap)
	copy(newStack, vm.stack)
	vm.stack = newStack
	vm.stackCap = newCap
	return nil
}

func (vm *VM) pop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.FLOAT_OBJ && rightType == object.FLOAT_OBJ:
		return vm.executeBinaryFloatOperation(op, left, right)
	case leftType == object.INTEGER_OBJ && rightType == object.FLOAT_OBJ:
		convertedLeft := &object.Float{Value: float64(left.(*object.Integer).Value)}
		return vm.executeBinaryFloatOperation(op, convertedLeft, right)
	case leftType == object.FLOAT_OBJ && rightType == object.INTEGER_OBJ:
		convertedRight := &object.Float{Value: float64(right.(*object.Integer).Value)}
		return vm.executeBinaryFloatOperation(op, left, convertedRight)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	if left == nil || right == nil {
		return fmt.Errorf("nil operand in binary integer operation")
	}
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}

	vm.push(object.NewInteger(result))
	return nil
}

func (vm *VM) executeBinaryFloatOperation(op code.Opcode, left, right object.Object) error {
	if left == nil || right == nil {
		return fmt.Errorf("nil operand in binary float operation")
	}
	leftValue := left.(*object.Float).Value
	rightValue := right.(*object.Float).Value
	var result float64
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
	vm.push(&object.Float{Value: result})
	return nil
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	var result string
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}

	vm.push(&object.String{Value: result})
	return nil
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	if left == nil || right == nil {
		return fmt.Errorf("nil operand in integer comparison")
	}
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result bool
	switch op {
	case code.OpEqual:
		result = leftValue == rightValue
	case code.OpNotEqual:
		result = leftValue != rightValue
	case code.OpGreaterThan:
		result = leftValue > rightValue
	case code.OpLessThan:
		result = leftValue < rightValue
	case code.OpGreaterThanEqual:
		result = leftValue >= rightValue
	case code.OpLessThanEqual:
		result = leftValue <= rightValue
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}

	return vm.push(nativeBoolToBooleanObject(result))
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case object.TRUE:
		return vm.push(object.FALSE)
	case object.FALSE:
		return vm.push(object.TRUE)
	case object.NULL:
		return vm.push(object.TRUE)
	default:
		return vm.push(object.FALSE)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	if operand == nil {
		return fmt.Errorf("nil operand in minus operation")
	}
	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
	value := operand.(*object.Integer).Value
	return vm.push(object.NewInteger(-value))
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		pairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: pairs}, nil
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	if array == nil || index == nil {
		return fmt.Errorf("nil operand in array index operation")
	}
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(object.NULL)
	}
	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	if hash == nil || index == nil {
		return fmt.Errorf("nil operand in hash index operation")
	}
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(object.NULL)
	}
	return vm.push(pair.Value)
}

func (vm *VM) currentFrame() *Frame {
	if vm.frameIndex == 0 {
		return nil
	}
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

// func (vm *VM) callFunction(fn *object.CompiledFunction, numArgs int) error {
// 	fn, ok := vm.stack[vm.sp-1-numArgs].(*object.CompiledFunction)
// 	if !ok {
// 		return fmt.Errorf("calling non-function")
// 	}
// 	if numArgs != fn.NumParameters {
// 		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", fn.NumParameters, numArgs)
// 	}
// 	frame := NewFrame(fn, vm.sp-numArgs)
// 	vm.pushFrame(frame)
// 	vm.sp = frame.basePointer + fn.NumLocals
// 	return nil
// }

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-builtin")
	}
}

func (vm *VM) callBuiltin(fn *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := fn.Fn(args...)
	vm.sp = vm.sp - numArgs - 1
	if result != nil {
		vm.push(result)
	} else {
		vm.push(object.NULL)
	}
	return nil
}

func (vm *VM) callClosure(closure *object.Closure, numArgs int) error {
	if numArgs != closure.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", closure.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(closure, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + closure.Fn.NumLocals
	return nil
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}
	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, FreeVariables: free}
	return vm.push(closure)
}
