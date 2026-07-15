package vm

import (
	"fmt"
	"monkey/object"
)

// ValueType represents the type of a tagged value
type ValueType byte

const (
	ValNil ValueType = iota
	ValBool
	ValInt
	ValFloat
	ValObject // pointer to object.Object (String, Array, Hash, Function, etc.)
)

// Value is a tagged union that avoids interface{} overhead for primitives
type Value struct {
	typ  ValueType
	ival int64         // used for ValInt and ValBool (0=false, 1=true)
	fval float64       // used for ValFloat
	obj  object.Object // used for ValObject (GC-visible)
}

// --- Constructors ---

func NilVal() Value {
	return Value{typ: ValNil}
}

func BoolVal(b bool) Value {
	v := Value{typ: ValBool}
	if b {
		v.ival = 1
	}
	return v
}

func IntVal(i int64) Value {
	return Value{typ: ValInt, ival: i}
}

func FloatVal(f float64) Value {
	return Value{typ: ValFloat, fval: f}
}

func ObjectVal(o object.Object) Value {
	return Value{typ: ValObject, obj: o}
}

// --- Type Checks ---

func (v Value) IsNil() bool     { return v.typ == ValNil }
func (v Value) IsBool() bool    { return v.typ == ValBool }
func (v Value) IsInt() bool     { return v.typ == ValInt }
func (v Value) IsFloat() bool   { return v.typ == ValFloat }
func (v Value) IsObject() bool  { return v.typ == ValObject }
func (v Value) Type() ValueType { return v.typ }

// --- Value Extraction ---

func (v Value) AsInt() int64            { return v.ival }
func (v Value) AsFloat() float64        { return v.fval }
func (v Value) AsBool() bool            { return v.ival != 0 }
func (v Value) AsObject() object.Object { return v.obj }

// --- Conversion to/from object.Object ---

func FromObject(o object.Object) Value {
	if o == nil {
		return NilVal()
	}
	switch obj := o.(type) {
	case *object.Integer:
		return IntVal(obj.Value)
	case *object.Float:
		return FloatVal(obj.Value)
	case *object.Boolean:
		return BoolVal(obj.Value)
	case *object.Null:
		return NilVal()
	default:
		return ObjectVal(o)
	}
}

func (v Value) ToObject() object.Object {
	switch v.typ {
	case ValNil:
		return object.NULL
	case ValBool:
		if v.ival != 0 {
			return object.TRUE
		}
		return object.FALSE
	case ValInt:
		return object.NewInteger(v.ival)
	case ValFloat:
		return &object.Float{Value: v.fval}
	case ValObject:
		return v.obj
	default:
		return object.NULL
	}
}

// --- Truthiness ---

func (v Value) IsTruthy() bool {
	switch v.typ {
	case ValNil:
		return false
	case ValBool:
		return v.ival != 0
	case ValInt:
		return true // all integers are truthy (including 0, matching Monkey semantics)
	case ValFloat:
		return true
	case ValObject:
		return v.obj != nil
	default:
		return false
	}
}

// --- String representation (for debugging) ---

func (v Value) String() string {
	switch v.typ {
	case ValNil:
		return "nil"
	case ValBool:
		if v.ival != 0 {
			return "true"
		}
		return "false"
	case ValInt:
		return fmt.Sprintf("%d", v.ival)
	case ValFloat:
		return fmt.Sprintf("%g", v.fval)
	case ValObject:
		if v.obj != nil {
			return v.obj.Inspect()
		}
		return "nil"
	default:
		return "<unknown>"
	}
}
