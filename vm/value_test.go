package vm

import (
	"math"
	"monkey/object"
	"testing"
)

func TestValueConstructors(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		typ      ValueType
		expected interface{}
	}{
		{"NilVal", NilVal(), ValNil, nil},
		{"BoolVal true", BoolVal(true), ValBool, true},
		{"BoolVal false", BoolVal(false), ValBool, false},
		{"IntVal positive", IntVal(42), ValInt, int64(42)},
		{"IntVal negative", IntVal(-100), ValInt, int64(-100)},
		{"IntVal zero", IntVal(0), ValInt, int64(0)},
		{"FloatVal positive", FloatVal(3.14), ValFloat, 3.14},
		{"FloatVal negative", FloatVal(-2.5), ValFloat, -2.5},
		{"FloatVal zero", FloatVal(0.0), ValFloat, 0.0},
		{"ObjectVal string", ObjectVal(&object.String{Value: "hello"}), ValObject, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.Type() != tt.typ {
				t.Errorf("wrong type. expected=%v, got=%v", tt.typ, tt.value.Type())
			}

			switch tt.typ {
			case ValNil:
				if !tt.value.IsNil() {
					t.Errorf("IsNil() should return true")
				}
			case ValBool:
				if !tt.value.IsBool() {
					t.Errorf("IsBool() should return true")
				}
				if tt.value.AsBool() != tt.expected.(bool) {
					t.Errorf("AsBool() = %v, want %v", tt.value.AsBool(), tt.expected)
				}
			case ValInt:
				if !tt.value.IsInt() {
					t.Errorf("IsInt() should return true")
				}
				if tt.value.AsInt() != tt.expected.(int64) {
					t.Errorf("AsInt() = %v, want %v", tt.value.AsInt(), tt.expected)
				}
			case ValFloat:
				if !tt.value.IsFloat() {
					t.Errorf("IsFloat() should return true")
				}
				if tt.value.AsFloat() != tt.expected.(float64) {
					t.Errorf("AsFloat() = %v, want %v", tt.value.AsFloat(), tt.expected)
				}
			case ValObject:
				if !tt.value.IsObject() {
					t.Errorf("IsObject() should return true")
				}
				str := tt.value.AsObject().(*object.String)
				if str.Value != tt.expected.(string) {
					t.Errorf("AsObject() = %v, want %v", str.Value, tt.expected)
				}
			}
		})
	}
}

func TestValueBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{"int64 max", IntVal(math.MaxInt64)},
		{"int64 min", IntVal(math.MinInt64)},
		{"float64 max", FloatVal(math.MaxFloat64)},
		{"float64 smallest positive", FloatVal(math.SmallestNonzeroFloat64)},
		{"float64 inf", FloatVal(math.Inf(1))},
		{"float64 -inf", FloatVal(math.Inf(-1))},
		{"float64 NaN", FloatVal(math.NaN())},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic and should roundtrip through ToObject()
			obj := tt.value.ToObject()
			if obj == nil {
				t.Errorf("ToObject() returned nil")
			}
		})
	}
}

func TestFromObjectToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    object.Object
		expected object.Object
	}{
		{"nil", nil, object.NULL},
		{"NULL", object.NULL, object.NULL},
		{"TRUE", object.TRUE, object.TRUE},
		{"FALSE", object.FALSE, object.FALSE},
		{"Integer", &object.Integer{Value: 42}, &object.Integer{Value: 42}},
		{"Float", &object.Float{Value: 3.14}, &object.Float{Value: 3.14}},
		{"String", &object.String{Value: "hello"}, &object.String{Value: "hello"}},
		{"Array", &object.Array{Elements: []object.Object{}}, &object.Array{Elements: []object.Object{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := FromObject(tt.input)
			result := value.ToObject()

			switch expected := tt.expected.(type) {
			case *object.Null:
				if result != object.NULL {
					t.Errorf("expected NULL, got %v", result)
				}
			case *object.Boolean:
				if result != expected {
					t.Errorf("expected %v, got %v", expected, result)
				}
			case *object.Integer:
				resultInt := result.(*object.Integer)
				if resultInt.Value != expected.Value {
					t.Errorf("expected %d, got %d", expected.Value, resultInt.Value)
				}
			case *object.Float:
				resultFloat := result.(*object.Float)
				if resultFloat.Value != expected.Value {
					t.Errorf("expected %f, got %f", expected.Value, resultFloat.Value)
				}
			case *object.String:
				resultStr := result.(*object.String)
				if resultStr.Value != expected.Value {
					t.Errorf("expected %s, got %s", expected.Value, resultStr.Value)
				}
			case *object.Array:
				if result.Type() != object.ARRAY_OBJ {
					t.Errorf("expected ARRAY_OBJ, got %v", result.Type())
				}
			}
		})
	}
}

func TestValueIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		{"NilVal", NilVal(), false},
		{"BoolVal true", BoolVal(true), true},
		{"BoolVal false", BoolVal(false), false},
		{"IntVal zero", IntVal(0), true}, // Monkey semantics: 0 is truthy
		{"IntVal positive", IntVal(1), true},
		{"IntVal negative", IntVal(-1), true},
		{"FloatVal zero", FloatVal(0.0), true},
		{"FloatVal positive", FloatVal(1.5), true},
		{"ObjectVal string", ObjectVal(&object.String{Value: ""}), true},
		{"ObjectVal nil", ObjectVal(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.IsTruthy() != tt.expected {
				t.Errorf("IsTruthy() = %v, want %v", tt.value.IsTruthy(), tt.expected)
			}
		})
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"NilVal", NilVal(), "nil"},
		{"BoolVal true", BoolVal(true), "true"},
		{"BoolVal false", BoolVal(false), "false"},
		{"IntVal", IntVal(42), "42"},
		{"FloatVal", FloatVal(3.14), "3.14"},
		{"ObjectVal string", ObjectVal(&object.String{Value: "hello"}), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValueTypeChecks(t *testing.T) {
	nilVal := NilVal()
	boolVal := BoolVal(true)
	intVal := IntVal(42)
	floatVal := FloatVal(3.14)
	objVal := ObjectVal(&object.String{Value: "test"})

	// Test IsNil
	if !nilVal.IsNil() {
		t.Error("nilVal.IsNil() should be true")
	}
	if boolVal.IsNil() || intVal.IsNil() || floatVal.IsNil() || objVal.IsNil() {
		t.Error("non-nil values should return false for IsNil()")
	}

	// Test IsBool
	if !boolVal.IsBool() {
		t.Error("boolVal.IsBool() should be true")
	}
	if nilVal.IsBool() || intVal.IsBool() || floatVal.IsBool() || objVal.IsBool() {
		t.Error("non-bool values should return false for IsBool()")
	}

	// Test IsInt
	if !intVal.IsInt() {
		t.Error("intVal.IsInt() should be true")
	}
	if nilVal.IsInt() || boolVal.IsInt() || floatVal.IsInt() || objVal.IsInt() {
		t.Error("non-int values should return false for IsInt()")
	}

	// Test IsFloat
	if !floatVal.IsFloat() {
		t.Error("floatVal.IsFloat() should be true")
	}
	if nilVal.IsFloat() || boolVal.IsFloat() || intVal.IsFloat() || objVal.IsFloat() {
		t.Error("non-float values should return false for IsFloat()")
	}

	// Test IsObject
	if !objVal.IsObject() {
		t.Error("objVal.IsObject() should be true")
	}
	if nilVal.IsObject() || boolVal.IsObject() || intVal.IsObject() || floatVal.IsObject() {
		t.Error("non-object values should return false for IsObject()")
	}
}
