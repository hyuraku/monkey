package object

import (
	"regexp"
	"strings"
	"testing"
)

func TestStringHashkey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestSplitBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal splitting cases
		{
			args:     []Object{&String{Value: "a,b,c"}, &String{Value: ","}},
			expected: []string{"a", "b", "c"},
		},
		{
			args:     []Object{&String{Value: "hello world test"}, &String{Value: " "}},
			expected: []string{"hello", "world", "test"},
		},
		{
			args:     []Object{&String{Value: "one::two::three"}, &String{Value: "::"}},
			expected: []string{"one", "two", "three"},
		},
		{
			args:     []Object{&String{Value: "a|b|c|d"}, &String{Value: "|"}},
			expected: []string{"a", "b", "c", "d"},
		},
		// Edge case: empty string
		{
			args:     []Object{&String{Value: ""}, &String{Value: ","}},
			expected: []string{""},
		},
		// Edge case: delimiter not found
		{
			args:     []Object{&String{Value: "abc"}, &String{Value: ","}},
			expected: []string{"abc"},
		},
		// Edge case: string starts with delimiter
		{
			args:     []Object{&String{Value: ",a,b,c"}, &String{Value: ","}},
			expected: []string{"", "a", "b", "c"},
		},
		// Edge case: string ends with delimiter
		{
			args:     []Object{&String{Value: "a,b,c,"}, &String{Value: ","}},
			expected: []string{"a", "b", "c", ""},
		},
		// Edge case: consecutive delimiters
		{
			args:     []Object{&String{Value: "a,,b"}, &String{Value: ","}},
			expected: []string{"a", "", "b"},
		},
		// Edge case: only delimiter
		{
			args:     []Object{&String{Value: ","}, &String{Value: ","}},
			expected: []string{"", ""},
		},
		// Edge case: multiple character delimiter
		{
			args:     []Object{&String{Value: "hello<-->world<-->test"}, &String{Value: "<-->"}},
			expected: []string{"hello", "world", "test"},
		},
		// Error case: empty delimiter
		{
			args:     []Object{&String{Value: "abc"}, &String{Value: ""}},
			expected: "delimiter cannot be empty",
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{&String{Value: "abc"}},
			expected: "wrong number of arguments. got=1, want=2",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&String{Value: "abc"}, &String{Value: ","}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: first argument not string
		{
			args:     []Object{&Integer{Value: 123}, &String{Value: ","}},
			expected: "first argument to `split` must be STRING, got INTEGER",
		},
		// Error case: second argument not string
		{
			args:     []Object{&String{Value: "abc"}, &Integer{Value: 123}},
			expected: "second argument to `split` must be STRING, got INTEGER",
		},
		// Error case: nil first argument
		{
			args:     []Object{nil, &String{Value: ","}},
			expected: "first argument to `split` cannot be nil",
		},
		// Error case: nil second argument
		{
			args:     []Object{&String{Value: "abc"}, nil},
			expected: "second argument to `split` cannot be nil",
		},
	}

	splitBuiltin := GetBuiltinByName("split")
	if splitBuiltin == nil {
		t.Fatal("split builtin not found")
	}

	for i, tt := range tests {
		result := splitBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case []string:
			// Test successful split result
			arr, ok := result.(*Array)
			if !ok {
				t.Errorf("test %d: expected Array, got %T (%+v)", i, result, result)
				continue
			}

			if len(arr.Elements) != len(expected) {
				t.Errorf("test %d: expected %d elements, got %d", i, len(expected), len(arr.Elements))
				continue
			}

			for j, elem := range arr.Elements {
				str, ok := elem.(*String)
				if !ok {
					t.Errorf("test %d: element %d is not String, got %T", i, j, elem)
					continue
				}
				if str.Value != expected[j] {
					t.Errorf("test %d: element %d expected %q, got %q", i, j, expected[j], str.Value)
				}
			}

		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}

func TestJoinBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal joining cases
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}, &String{Value: "c"}}},
				&String{Value: ","},
			},
			expected: "a,b,c",
		},
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "hello"}, &String{Value: "world"}, &String{Value: "test"}}},
				&String{Value: " "},
			},
			expected: "hello world test",
		},
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "one"}, &String{Value: "two"}, &String{Value: "three"}}},
				&String{Value: "::"},
			},
			expected: "one::two::three",
		},
		// Mixed types - should use Inspect() method
		{
			args: []Object{
				&Array{Elements: []Object{&Integer{Value: 1}, &String{Value: "hello"}, &Boolean{Value: true}}},
				&String{Value: ","},
			},
			expected: "1,hello,true",
		},
		{
			args: []Object{
				&Array{Elements: []Object{&Integer{Value: 42}, &Float{Value: 3.14}, &String{Value: "test"}}},
				&String{Value: "|"},
			},
			expected: "42|3.140000|test",
		},
		// Edge case: empty array
		{
			args: []Object{
				&Array{Elements: []Object{}},
				&String{Value: ","},
			},
			expected: "",
		},
		// Edge case: single element
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "single"}}},
				&String{Value: ","},
			},
			expected: "single",
		},
		// Edge case: empty delimiter
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}, &String{Value: "b"}}},
				&String{Value: ""},
			},
			expected: "ab",
		},
		// Edge case: array with nil elements
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}, nil, &String{Value: "c"}}},
				&String{Value: ","},
			},
			expected: "a,,c",
		},
		// Edge case: array with only nil elements
		{
			args: []Object{
				&Array{Elements: []Object{nil, nil}},
				&String{Value: ","},
			},
			expected: ",",
		},
		// Edge case: complex objects (Hash, Array)
		{
			args: []Object{
				&Array{Elements: []Object{
					&Array{Elements: []Object{&Integer{Value: 1}, &Integer{Value: 2}}},
					&String{Value: "middle"},
					&Hash{Pairs: map[HashKey]HashPair{
						(&String{Value: "key"}).HashKey(): {Key: &String{Value: "key"}, Value: &String{Value: "value"}},
					}},
				}},
				&String{Value: " | "},
			},
			expected: "[1, 2] | middle | {key: value}",
		},
		// Error case: wrong number of arguments - too few
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}}},
			},
			expected: "wrong number of arguments. got=1, want=2",
		},
		// Error case: wrong number of arguments - too many
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}}},
				&String{Value: ","},
				&String{Value: "extra"},
			},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: first argument not array
		{
			args: []Object{
				&String{Value: "not-an-array"},
				&String{Value: ","},
			},
			expected: "first argument to `join` must be ARRAY, got STRING",
		},
		// Error case: second argument not string
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}}},
				&Integer{Value: 42},
			},
			expected: "second argument to `join` must be STRING, got INTEGER",
		},
		// Error case: nil first argument
		{
			args: []Object{
				nil,
				&String{Value: ","},
			},
			expected: "first argument to `join` cannot be nil",
		},
		// Error case: nil second argument
		{
			args: []Object{
				&Array{Elements: []Object{&String{Value: "a"}}},
				nil,
			},
			expected: "second argument to `join` cannot be nil",
		},
		// Error case: no arguments
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=2",
		},
	}

	joinBuiltin := GetBuiltinByName("join")
	if joinBuiltin == nil {
		t.Fatal("join builtin not found")
	}

	for i, tt := range tests {
		result := joinBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case string:
			// Check if this is an error message
			isError := (len(expected) >= 5 && expected[:5] == "wrong") ||
				(len(expected) >= 5 && expected[:5] == "first") ||
				(len(expected) >= 6 && expected[:6] == "second")

			if isError {
				// Test error result
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			} else {
				// Test successful join result
				str, ok := result.(*String)
				if !ok {
					t.Errorf("test %d: expected String, got %T (%+v)", i, result, result)
					continue
				}
				if str.Value != expected {
					t.Errorf("test %d: expected %q, got %q", i, expected, str.Value)
				}
			}
		}
	}
}

func TestUpperBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal case: lowercase to uppercase
		{
			args:     []Object{&String{Value: "hello"}},
			expected: "HELLO",
		},
		{
			args:     []Object{&String{Value: "world"}},
			expected: "WORLD",
		},
		// Mixed case
		{
			args:     []Object{&String{Value: "Hello World"}},
			expected: "HELLO WORLD",
		},
		{
			args:     []Object{&String{Value: "mOnKeY"}},
			expected: "MONKEY",
		},
		// Empty string
		{
			args:     []Object{&String{Value: ""}},
			expected: "",
		},
		// Numbers and symbols
		{
			args:     []Object{&String{Value: "hello123!"}},
			expected: "HELLO123!",
		},
		{
			args:     []Object{&String{Value: "test@example.com"}},
			expected: "TEST@EXAMPLE.COM",
		},
		{
			args:     []Object{&String{Value: "123!@#$%"}},
			expected: "123!@#$%",
		},
		// Already uppercase
		{
			args:     []Object{&String{Value: "HELLO"}},
			expected: "HELLO",
		},
		{
			args:     []Object{&String{Value: "ALREADY UPPERCASE"}},
			expected: "ALREADY UPPERCASE",
		},
		// Special characters and Unicode
		{
			args:     []Object{&String{Value: "café"}},
			expected: "CAFÉ",
		},
		{
			args:     []Object{&String{Value: "naïve"}},
			expected: "NAÏVE",
		},
		// Single character
		{
			args:     []Object{&String{Value: "a"}},
			expected: "A",
		},
		{
			args:     []Object{&String{Value: "Z"}},
			expected: "Z",
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=1",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&String{Value: "hello"}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=2, want=1",
		},
		// Error case: non-string argument
		{
			args:     []Object{&Integer{Value: 123}},
			expected: "argument to `upper` must be STRING, got INTEGER",
		},
		{
			args:     []Object{&Boolean{Value: true}},
			expected: "argument to `upper` must be STRING, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&String{Value: "hello"}}}},
			expected: "argument to `upper` must be STRING, got ARRAY",
		},
		// Error case: nil argument
		{
			args:     []Object{nil},
			expected: "argument to `upper` cannot be nil",
		},
	}

	upperBuiltin := GetBuiltinByName("upper")
	if upperBuiltin == nil {
		t.Fatal("upper builtin not found")
	}

	for i, tt := range tests {
		result := upperBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case string:
			if (len(expected) >= 5 && expected[:5] == "wrong") || (len(expected) >= 8 && expected[:8] == "argument") {
				// Test error result
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			} else {
				// Test successful upper result
				str, ok := result.(*String)
				if !ok {
					t.Errorf("test %d: expected String, got %T (%+v)", i, result, result)
					continue
				}
				if str.Value != expected {
					t.Errorf("test %d: expected %q, got %q", i, expected, str.Value)
				}
			}
		}
	}
}

func TestLowerBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal case: uppercase to lowercase
		{
			args:     []Object{&String{Value: "WORLD"}},
			expected: "world",
		},
		{
			args:     []Object{&String{Value: "HELLO"}},
			expected: "hello",
		},
		// Mixed case
		{
			args:     []Object{&String{Value: "Hello World"}},
			expected: "hello world",
		},
		{
			args:     []Object{&String{Value: "MoNkEy"}},
			expected: "monkey",
		},
		// Empty string
		{
			args:     []Object{&String{Value: ""}},
			expected: "",
		},
		// Numbers and symbols
		{
			args:     []Object{&String{Value: "HELLO123!"}},
			expected: "hello123!",
		},
		{
			args:     []Object{&String{Value: "TEST@EXAMPLE.COM"}},
			expected: "test@example.com",
		},
		{
			args:     []Object{&String{Value: "123!@#$%"}},
			expected: "123!@#$%",
		},
		// Already lowercase
		{
			args:     []Object{&String{Value: "hello"}},
			expected: "hello",
		},
		{
			args:     []Object{&String{Value: "already lowercase"}},
			expected: "already lowercase",
		},
		// Special characters and Unicode
		{
			args:     []Object{&String{Value: "CAFÉ"}},
			expected: "café",
		},
		{
			args:     []Object{&String{Value: "NAÏVE"}},
			expected: "naïve",
		},
		// Single character
		{
			args:     []Object{&String{Value: "A"}},
			expected: "a",
		},
		{
			args:     []Object{&String{Value: "z"}},
			expected: "z",
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=1",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&String{Value: "HELLO"}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=2, want=1",
		},
		// Error case: non-string argument
		{
			args:     []Object{&Integer{Value: 123}},
			expected: "argument to `lower` must be STRING, got INTEGER",
		},
		{
			args:     []Object{&Boolean{Value: true}},
			expected: "argument to `lower` must be STRING, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&String{Value: "HELLO"}}}},
			expected: "argument to `lower` must be STRING, got ARRAY",
		},
		// Error case: nil argument
		{
			args:     []Object{nil},
			expected: "argument to `lower` cannot be nil",
		},
	}

	lowerBuiltin := GetBuiltinByName("lower")
	if lowerBuiltin == nil {
		t.Fatal("lower builtin not found")
	}

	for i, tt := range tests {
		result := lowerBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case string:
			if (len(expected) >= 5 && expected[:5] == "wrong") || (len(expected) >= 8 && expected[:8] == "argument") {
				// Test error result
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			} else {
				// Test successful lower result
				str, ok := result.(*String)
				if !ok {
					t.Errorf("test %d: expected String, got %T (%+v)", i, result, result)
					continue
				}
				if str.Value != expected {
					t.Errorf("test %d: expected %q, got %q", i, expected, str.Value)
				}
			}
		}
	}
}

func TestAbsBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: positive integers
		{
			args:     []Object{&Integer{Value: 5}},
			expected: 5,
		},
		{
			args:     []Object{&Integer{Value: 42}},
			expected: 42,
		},
		// Normal cases: negative integers
		{
			args:     []Object{&Integer{Value: -5}},
			expected: 5,
		},
		{
			args:     []Object{&Integer{Value: -42}},
			expected: 42,
		},
		// Normal cases: positive floats
		{
			args:     []Object{&Float{Value: 3.14}},
			expected: 3.14,
		},
		{
			args:     []Object{&Float{Value: 2.5}},
			expected: 2.5,
		},
		// Normal cases: negative floats
		{
			args:     []Object{&Float{Value: -3.14}},
			expected: 3.14,
		},
		{
			args:     []Object{&Float{Value: -2.5}},
			expected: 2.5,
		},
		// Zero cases
		{
			args:     []Object{&Integer{Value: 0}},
			expected: 0,
		},
		{
			args:     []Object{&Float{Value: 0.0}},
			expected: 0.0,
		},
		{
			args:     []Object{&Float{Value: -0.0}},
			expected: 0.0,
		},
		// Edge cases: very large numbers
		{
			args:     []Object{&Integer{Value: -9223372036854775807}},
			expected: 9223372036854775807,
		},
		{
			args:     []Object{&Float{Value: -999999.999999}},
			expected: 999999.999999,
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=1",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&Integer{Value: 5}, &Integer{Value: 10}},
			expected: "wrong number of arguments. got=2, want=1",
		},
		// Error case: non-numeric argument
		{
			args:     []Object{&String{Value: "not a number"}},
			expected: "argument to `abs` must be INTEGER or FLOAT, got STRING",
		},
		{
			args:     []Object{&Boolean{Value: true}},
			expected: "argument to `abs` must be INTEGER or FLOAT, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&Integer{Value: 5}}}},
			expected: "argument to `abs` must be INTEGER or FLOAT, got ARRAY",
		},
		// Error case: nil argument
		{
			args:     []Object{nil},
			expected: "argument to `abs` cannot be nil",
		},
	}

	absBuiltin := GetBuiltinByName("abs")
	if absBuiltin == nil {
		t.Fatal("abs builtin not found")
	}

	for i, tt := range tests {
		result := absBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case int64:
			// Test successful integer result
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != expected {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case int:
			// Test successful integer result (handle int literals)
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != int64(expected) {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case float64:
			// Test successful float result
			floatObj, ok := result.(*Float)
			if !ok {
				t.Errorf("test %d: expected Float, got %T (%+v)", i, result, result)
				continue
			}
			if floatObj.Value != expected {
				t.Errorf("test %d: expected %f, got %f", i, expected, floatObj.Value)
			}
		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}

func TestMinBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: integers
		{
			args:     []Object{&Integer{Value: 3}, &Integer{Value: 7}},
			expected: 3,
		},
		{
			args:     []Object{&Integer{Value: 7}, &Integer{Value: 3}},
			expected: 3,
		},
		{
			args:     []Object{&Integer{Value: -5}, &Integer{Value: 10}},
			expected: -5,
		},
		{
			args:     []Object{&Integer{Value: 0}, &Integer{Value: 5}},
			expected: 0,
		},
		// Normal cases: floats
		{
			args:     []Object{&Float{Value: 3.14}, &Float{Value: 2.71}},
			expected: 2.71,
		},
		{
			args:     []Object{&Float{Value: -1.5}, &Float{Value: 2.5}},
			expected: -1.5,
		},
		{
			args:     []Object{&Float{Value: 0.0}, &Float{Value: 0.1}},
			expected: 0.0,
		},
		// Mixed int/float cases
		{
			args:     []Object{&Integer{Value: 3}, &Float{Value: 2.5}},
			expected: 2.5,
		},
		{
			args:     []Object{&Float{Value: 3.14}, &Integer{Value: 4}},
			expected: 3.14,
		},
		{
			args:     []Object{&Integer{Value: -2}, &Float{Value: -1.5}},
			expected: -2.0,
		},
		{
			args:     []Object{&Float{Value: 5.0}, &Integer{Value: 5}},
			expected: 5.0,
		},
		// Equal values
		{
			args:     []Object{&Integer{Value: 5}, &Integer{Value: 5}},
			expected: 5,
		},
		{
			args:     []Object{&Float{Value: 3.14}, &Float{Value: 3.14}},
			expected: 3.14,
		},
		// Zero cases
		{
			args:     []Object{&Integer{Value: 0}, &Integer{Value: 0}},
			expected: 0,
		},
		{
			args:     []Object{&Float{Value: 0.0}, &Float{Value: -0.0}},
			expected: 0.0,
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{&Integer{Value: 5}},
			expected: "wrong number of arguments. got=1, want=2",
		},
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=2",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: non-numeric arguments
		{
			args:     []Object{&String{Value: "not a number"}, &Integer{Value: 5}},
			expected: "first argument to `min` must be INTEGER or FLOAT, got STRING",
		},
		{
			args:     []Object{&Integer{Value: 5}, &Boolean{Value: true}},
			expected: "second argument to `min` must be INTEGER or FLOAT, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&Integer{Value: 5}}}, &Integer{Value: 3}},
			expected: "first argument to `min` must be INTEGER or FLOAT, got ARRAY",
		},
		// Error case: nil arguments
		{
			args:     []Object{nil, &Integer{Value: 5}},
			expected: "first argument to `min` cannot be nil",
		},
		{
			args:     []Object{&Integer{Value: 5}, nil},
			expected: "second argument to `min` cannot be nil",
		},
	}

	minBuiltin := GetBuiltinByName("min")
	if minBuiltin == nil {
		t.Fatal("min builtin not found")
	}

	for i, tt := range tests {
		result := minBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case int64:
			// Test successful integer result
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != expected {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case int:
			// Test successful integer result (handle int literals)
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != int64(expected) {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case float64:
			// Test successful float result
			floatObj, ok := result.(*Float)
			if !ok {
				t.Errorf("test %d: expected Float, got %T (%+v)", i, result, result)
				continue
			}
			if floatObj.Value != expected {
				t.Errorf("test %d: expected %f, got %f", i, expected, floatObj.Value)
			}
		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}

func TestMaxBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: integers
		{
			args:     []Object{&Integer{Value: 3}, &Integer{Value: 7}},
			expected: 7,
		},
		{
			args:     []Object{&Integer{Value: 7}, &Integer{Value: 3}},
			expected: 7,
		},
		{
			args:     []Object{&Integer{Value: -5}, &Integer{Value: 10}},
			expected: 10,
		},
		{
			args:     []Object{&Integer{Value: 0}, &Integer{Value: 5}},
			expected: 5,
		},
		// Normal cases: floats
		{
			args:     []Object{&Float{Value: 3.14}, &Float{Value: 2.71}},
			expected: 3.14,
		},
		{
			args:     []Object{&Float{Value: -1.5}, &Float{Value: 2.5}},
			expected: 2.5,
		},
		{
			args:     []Object{&Float{Value: 0.0}, &Float{Value: 0.1}},
			expected: 0.1,
		},
		// Mixed int/float cases
		{
			args:     []Object{&Integer{Value: 3}, &Float{Value: 2.5}},
			expected: 3.0,
		},
		{
			args:     []Object{&Float{Value: 3.14}, &Integer{Value: 4}},
			expected: 4.0,
		},
		{
			args:     []Object{&Integer{Value: -2}, &Float{Value: -1.5}},
			expected: -1.5,
		},
		{
			args:     []Object{&Float{Value: 5.0}, &Integer{Value: 5}},
			expected: 5.0,
		},
		// Equal values
		{
			args:     []Object{&Integer{Value: 5}, &Integer{Value: 5}},
			expected: 5,
		},
		{
			args:     []Object{&Float{Value: 3.14}, &Float{Value: 3.14}},
			expected: 3.14,
		},
		// Zero cases
		{
			args:     []Object{&Integer{Value: 0}, &Integer{Value: 0}},
			expected: 0,
		},
		{
			args:     []Object{&Float{Value: 0.0}, &Float{Value: -0.0}},
			expected: 0.0,
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{&Integer{Value: 5}},
			expected: "wrong number of arguments. got=1, want=2",
		},
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=2",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&Integer{Value: 1}, &Integer{Value: 2}, &Integer{Value: 3}},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: non-numeric arguments
		{
			args:     []Object{&String{Value: "not a number"}, &Integer{Value: 5}},
			expected: "first argument to `max` must be INTEGER or FLOAT, got STRING",
		},
		{
			args:     []Object{&Integer{Value: 5}, &Boolean{Value: true}},
			expected: "second argument to `max` must be INTEGER or FLOAT, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&Integer{Value: 5}}}, &Integer{Value: 3}},
			expected: "first argument to `max` must be INTEGER or FLOAT, got ARRAY",
		},
		// Error case: nil arguments
		{
			args:     []Object{nil, &Integer{Value: 5}},
			expected: "first argument to `max` cannot be nil",
		},
		{
			args:     []Object{&Integer{Value: 5}, nil},
			expected: "second argument to `max` cannot be nil",
		},
	}

	maxBuiltin := GetBuiltinByName("max")
	if maxBuiltin == nil {
		t.Fatal("max builtin not found")
	}

	for i, tt := range tests {
		result := maxBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case int64:
			// Test successful integer result
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != expected {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case int:
			// Test successful integer result (handle int literals)
			intObj, ok := result.(*Integer)
			if !ok {
				t.Errorf("test %d: expected Integer, got %T (%+v)", i, result, result)
				continue
			}
			if intObj.Value != int64(expected) {
				t.Errorf("test %d: expected %d, got %d", i, expected, intObj.Value)
			}
		case float64:
			// Test successful float result
			floatObj, ok := result.(*Float)
			if !ok {
				t.Errorf("test %d: expected Float, got %T (%+v)", i, result, result)
				continue
			}
			if floatObj.Value != expected {
				t.Errorf("test %d: expected %f, got %f", i, expected, floatObj.Value)
			}
		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}

func TestSqrtBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: perfect squares
		{
			args:     []Object{&Integer{Value: 9}},
			expected: 3.0,
		},
		{
			args:     []Object{&Integer{Value: 16}},
			expected: 4.0,
		},
		{
			args:     []Object{&Integer{Value: 25}},
			expected: 5.0,
		},
		{
			args:     []Object{&Integer{Value: 100}},
			expected: 10.0,
		},
		// Normal cases: non-perfect squares
		{
			args:     []Object{&Integer{Value: 2}},
			expected: 1.4142135623730951,
		},
		{
			args:     []Object{&Integer{Value: 8}},
			expected: 2.8284271247461903,
		},
		// Normal cases: floats
		{
			args:     []Object{&Float{Value: 2.25}},
			expected: 1.5,
		},
		{
			args:     []Object{&Float{Value: 6.25}},
			expected: 2.5,
		},
		{
			args:     []Object{&Float{Value: 0.25}},
			expected: 0.5,
		},
		// Zero cases
		{
			args:     []Object{&Integer{Value: 0}},
			expected: 0.0,
		},
		{
			args:     []Object{&Float{Value: 0.0}},
			expected: 0.0,
		},
		// One case
		{
			args:     []Object{&Integer{Value: 1}},
			expected: 1.0,
		},
		{
			args:     []Object{&Float{Value: 1.0}},
			expected: 1.0,
		},
		// Edge cases: large numbers
		{
			args:     []Object{&Integer{Value: 144}},
			expected: 12.0,
		},
		{
			args:     []Object{&Float{Value: 12.25}},
			expected: 3.5,
		},
		// Error case: negative numbers
		{
			args:     []Object{&Integer{Value: -9}},
			expected: "sqrt of negative number is not supported",
		},
		{
			args:     []Object{&Float{Value: -2.25}},
			expected: "sqrt of negative number is not supported",
		},
		{
			args:     []Object{&Integer{Value: -1}},
			expected: "sqrt of negative number is not supported",
		},
		// Error case: wrong number of arguments - too few
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=1",
		},
		// Error case: wrong number of arguments - too many
		{
			args:     []Object{&Integer{Value: 9}, &Integer{Value: 4}},
			expected: "wrong number of arguments. got=2, want=1",
		},
		// Error case: non-numeric arguments
		{
			args:     []Object{&String{Value: "not a number"}},
			expected: "argument to `sqrt` must be INTEGER or FLOAT, got STRING",
		},
		{
			args:     []Object{&Boolean{Value: true}},
			expected: "argument to `sqrt` must be INTEGER or FLOAT, got BOOLEAN",
		},
		{
			args:     []Object{&Array{Elements: []Object{&Integer{Value: 9}}}},
			expected: "argument to `sqrt` must be INTEGER or FLOAT, got ARRAY",
		},
		// Error case: nil argument
		{
			args:     []Object{nil},
			expected: "argument to `sqrt` cannot be nil",
		},
	}

	sqrtBuiltin := GetBuiltinByName("sqrt")
	if sqrtBuiltin == nil {
		t.Fatal("sqrt builtin not found")
	}

	for i, tt := range tests {
		result := sqrtBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case float64:
			// Test successful float result
			floatObj, ok := result.(*Float)
			if !ok {
				t.Errorf("test %d: expected Float, got %T (%+v)", i, result, result)
				continue
			}
			if floatObj.Value != expected {
				t.Errorf("test %d: expected %f, got %f", i, expected, floatObj.Value)
			}
		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}

func TestRegexBuiltin(t *testing.T) {
	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: valid regex patterns
		{
			args:     []Object{&String{Value: "hello"}},
			expected: "hello",
		},
		{
			args:     []Object{&String{Value: "\\d+"}},
			expected: "\\d+",
		},
		{
			args:     []Object{&String{Value: "[a-zA-Z]+"}},
			expected: "[a-zA-Z]+",
		},
		{
			args:     []Object{&String{Value: "^test$"}},
			expected: "^test$",
		},
		{
			args:     []Object{&String{Value: "foo.*bar"}},
			expected: "foo.*bar",
		},
		// Complex patterns
		{
			args:     []Object{&String{Value: "\\w+@\\w+\\.\\w+"}},
			expected: "\\w+@\\w+\\.\\w+",
		},
		{
			args:     []Object{&String{Value: "\\b\\w+\\s+\\w+\\b"}},
			expected: "\\b\\w+\\s+\\w+\\b",
		},
		// Error case: invalid regex patterns
		{
			args:     []Object{&String{Value: "["}},
			expected: "invalid regex pattern:",
		},
		{
			args:     []Object{&String{Value: "(unclosed"}},
			expected: "invalid regex pattern:",
		},
		{
			args:     []Object{&String{Value: "*"}},
			expected: "invalid regex pattern:",
		},
		// Error case: wrong number of arguments
		{
			args:     []Object{},
			expected: "wrong number of arguments. got=0, want=1",
		},
		{
			args:     []Object{&String{Value: "test"}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=2, want=1",
		},
		// Error case: non-string argument
		{
			args:     []Object{&Integer{Value: 123}},
			expected: "argument to `regex` must be STRING, got INTEGER",
		},
		{
			args:     []Object{&Boolean{Value: true}},
			expected: "argument to `regex` must be STRING, got BOOLEAN",
		},
		// Error case: nil argument
		{
			args:     []Object{nil},
			expected: "argument to `regex` cannot be nil",
		},
	}

	regexBuiltin := GetBuiltinByName("regex")
	if regexBuiltin == nil {
		t.Fatal("regex builtin not found")
	}

	for i, tt := range tests {
		result := regexBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case string:
			if expected == "invalid regex pattern:" {
				// Test error result for invalid patterns
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if !strings.Contains(errObj.Message, expected) {
					t.Errorf("test %d: expected error message containing %q, got %q", i, expected, errObj.Message)
				}
			} else if strings.HasPrefix(expected, "wrong number of arguments") || strings.HasPrefix(expected, "argument to") {
				// Test error result for argument errors
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			} else {
				// Test successful regex creation
				regexObj, ok := result.(*Regex)
				if !ok {
					t.Errorf("test %d: expected Regex, got %T (%+v)", i, result, result)
					continue
				}
				if regexObj.Pattern != expected {
					t.Errorf("test %d: expected pattern %q, got %q", i, expected, regexObj.Pattern)
				}
				if regexObj.Regexp == nil {
					t.Errorf("test %d: regex Regexp field is nil", i)
				}
			}
		}
	}
}

func TestMatchBuiltin(t *testing.T) {
	// Create test regex objects
	testRegex1, _ := regexp.Compile("hello")
	testRegex2, _ := regexp.Compile("\\d+")
	testRegex3, _ := regexp.Compile("(\\w+)@(\\w+)\\.com")
	testRegex4, _ := regexp.Compile("foo")

	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: successful matches
		{
			args:     []Object{&Regex{Pattern: "hello", Regexp: testRegex1}, &String{Value: "hello world"}},
			expected: []string{"hello"},
		},
		{
			args:     []Object{&Regex{Pattern: "\\d+", Regexp: testRegex2}, &String{Value: "test 123 end"}},
			expected: []string{"123"},
		},
		{
			args:     []Object{&Regex{Pattern: "(\\w+)@(\\w+)\\.com", Regexp: testRegex3}, &String{Value: "Contact us at john@example.com"}},
			expected: []string{"john@example.com", "john", "example"},
		},
		// No match cases
		{
			args:     []Object{&Regex{Pattern: "foo", Regexp: testRegex4}, &String{Value: "bar baz"}},
			expected: "null",
		},
		{
			args:     []Object{&Regex{Pattern: "\\d+", Regexp: testRegex2}, &String{Value: "no numbers here"}},
			expected: "null",
		},
		// Error case: wrong number of arguments
		{
			args:     []Object{&Regex{Pattern: "test", Regexp: testRegex1}},
			expected: "wrong number of arguments. got=1, want=2",
		},
		{
			args:     []Object{&Regex{Pattern: "test", Regexp: testRegex1}, &String{Value: "text"}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: wrong argument types
		{
			args:     []Object{&String{Value: "not regex"}, &String{Value: "text"}},
			expected: "first argument to `match` must be REGEX, got STRING",
		},
		{
			args:     []Object{&Regex{Pattern: "test", Regexp: testRegex1}, &Integer{Value: 123}},
			expected: "second argument to `match` must be STRING, got INTEGER",
		},
		// Error case: nil arguments
		{
			args:     []Object{nil, &String{Value: "text"}},
			expected: "first argument to `match` cannot be nil",
		},
		{
			args:     []Object{&Regex{Pattern: "test", Regexp: testRegex1}, nil},
			expected: "second argument to `match` cannot be nil",
		},
	}

	matchBuiltin := GetBuiltinByName("match")
	if matchBuiltin == nil {
		t.Fatal("match builtin not found")
	}

	for i, tt := range tests {
		result := matchBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case []string:
			// Test successful match result
			arr, ok := result.(*Array)
			if !ok {
				t.Errorf("test %d: expected Array, got %T (%+v)", i, result, result)
				continue
			}
			if len(arr.Elements) != len(expected) {
				t.Errorf("test %d: expected %d elements, got %d", i, len(expected), len(arr.Elements))
				continue
			}
			for j, elem := range arr.Elements {
				str, ok := elem.(*String)
				if !ok {
					t.Errorf("test %d: element %d is not String, got %T", i, j, elem)
					continue
				}
				if str.Value != expected[j] {
					t.Errorf("test %d: element %d expected %q, got %q", i, j, expected[j], str.Value)
				}
			}
		case string:
			if expected == "null" {
				// Test null result for no match
				if result != NULL {
					t.Errorf("test %d: expected NULL, got %T (%+v)", i, result, result)
				}
			} else {
				// Test error result
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			}
		}
	}
}

func TestReplaceBuiltin(t *testing.T) {
	// Create test regex objects
	testRegex1, _ := regexp.Compile("hello")
	testRegex2, _ := regexp.Compile("\\d+")
	testRegex3, _ := regexp.Compile("(\\w+)@(\\w+)\\.com")

	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: successful replacements
		{
			args:     []Object{&String{Value: "hello world"}, &Regex{Pattern: "hello", Regexp: testRegex1}, &String{Value: "hi"}},
			expected: "hi world",
		},
		{
			args:     []Object{&String{Value: "test 123 and 456"}, &Regex{Pattern: "\\d+", Regexp: testRegex2}, &String{Value: "X"}},
			expected: "test X and X",
		},
		{
			args:     []Object{&String{Value: "Email john@example.com"}, &Regex{Pattern: "(\\w+)@(\\w+)\\.com", Regexp: testRegex3}, &String{Value: "$1 at $2"}},
			expected: "Email john at example",
		},
		// No match cases
		{
			args:     []Object{&String{Value: "no match here"}, &Regex{Pattern: "\\d+", Regexp: testRegex2}, &String{Value: "X"}},
			expected: "no match here",
		},
		// Error case: wrong number of arguments
		{
			args:     []Object{&String{Value: "text"}, &Regex{Pattern: "test", Regexp: testRegex1}},
			expected: "wrong number of arguments. got=2, want=3",
		},
		{
			args:     []Object{&String{Value: "text"}, &Regex{Pattern: "test", Regexp: testRegex1}, &String{Value: "replacement"}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=4, want=3",
		},
		// Error case: wrong argument types
		{
			args:     []Object{&Integer{Value: 123}, &Regex{Pattern: "test", Regexp: testRegex1}, &String{Value: "replacement"}},
			expected: "first argument to `replace` must be STRING, got INTEGER",
		},
		{
			args:     []Object{&String{Value: "text"}, &String{Value: "not regex"}, &String{Value: "replacement"}},
			expected: "second argument to `replace` must be REGEX, got STRING",
		},
		{
			args:     []Object{&String{Value: "text"}, &Regex{Pattern: "test", Regexp: testRegex1}, &Integer{Value: 123}},
			expected: "third argument to `replace` must be STRING, got INTEGER",
		},
		// Error case: nil arguments
		{
			args:     []Object{nil, &Regex{Pattern: "test", Regexp: testRegex1}, &String{Value: "replacement"}},
			expected: "first argument to `replace` cannot be nil",
		},
		{
			args:     []Object{&String{Value: "text"}, nil, &String{Value: "replacement"}},
			expected: "second argument to `replace` cannot be nil",
		},
		{
			args:     []Object{&String{Value: "text"}, &Regex{Pattern: "test", Regexp: testRegex1}, nil},
			expected: "third argument to `replace` cannot be nil",
		},
	}

	replaceBuiltin := GetBuiltinByName("replace")
	if replaceBuiltin == nil {
		t.Fatal("replace builtin not found")
	}

	for i, tt := range tests {
		result := replaceBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case string:
			if strings.HasPrefix(expected, "wrong number of arguments") || strings.HasPrefix(expected, "first argument") || strings.HasPrefix(expected, "second argument") || strings.HasPrefix(expected, "third argument") {
				// Test error result
				errObj, ok := result.(*Error)
				if !ok {
					t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
					continue
				}
				if errObj.Message != expected {
					t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
				}
			} else {
				// Test successful replacement result
				str, ok := result.(*String)
				if !ok {
					t.Errorf("test %d: expected String, got %T (%+v)", i, result, result)
					continue
				}
				if str.Value != expected {
					t.Errorf("test %d: expected %q, got %q", i, expected, str.Value)
				}
			}
		}
	}
}

func TestRegexSplitBuiltin(t *testing.T) {
	// Create test regex objects
	testRegex1, _ := regexp.Compile(",")
	testRegex2, _ := regexp.Compile("\\s+")
	testRegex3, _ := regexp.Compile("\\d+")

	tests := []struct {
		args     []Object
		expected interface{}
	}{
		// Normal cases: successful splits
		{
			args:     []Object{&String{Value: "a,b,c"}, &Regex{Pattern: ",", Regexp: testRegex1}},
			expected: []string{"a", "b", "c"},
		},
		{
			args:     []Object{&String{Value: "hello   world    test"}, &Regex{Pattern: "\\s+", Regexp: testRegex2}},
			expected: []string{"hello", "world", "test"},
		},
		{
			args:     []Object{&String{Value: "abc123def456ghi"}, &Regex{Pattern: "\\d+", Regexp: testRegex3}},
			expected: []string{"abc", "def", "ghi"},
		},
		// Edge cases
		{
			args:     []Object{&String{Value: ""}, &Regex{Pattern: ",", Regexp: testRegex1}},
			expected: []string{""},
		},
		{
			args:     []Object{&String{Value: "no delimiters"}, &Regex{Pattern: ",", Regexp: testRegex1}},
			expected: []string{"no delimiters"},
		},
		// Error case: wrong number of arguments
		{
			args:     []Object{&String{Value: "text"}},
			expected: "wrong number of arguments. got=1, want=2",
		},
		{
			args:     []Object{&String{Value: "text"}, &Regex{Pattern: "test", Regexp: testRegex1}, &String{Value: "extra"}},
			expected: "wrong number of arguments. got=3, want=2",
		},
		// Error case: wrong argument types
		{
			args:     []Object{&Integer{Value: 123}, &Regex{Pattern: "test", Regexp: testRegex1}},
			expected: "first argument to `regex_split` must be STRING, got INTEGER",
		},
		{
			args:     []Object{&String{Value: "text"}, &String{Value: "not regex"}},
			expected: "second argument to `regex_split` must be REGEX, got STRING",
		},
		// Error case: nil arguments
		{
			args:     []Object{nil, &Regex{Pattern: "test", Regexp: testRegex1}},
			expected: "first argument to `regex_split` cannot be nil",
		},
		{
			args:     []Object{&String{Value: "text"}, nil},
			expected: "second argument to `regex_split` cannot be nil",
		},
	}

	regexSplitBuiltin := GetBuiltinByName("regex_split")
	if regexSplitBuiltin == nil {
		t.Fatal("regex_split builtin not found")
	}

	for i, tt := range tests {
		result := regexSplitBuiltin.Fn(tt.args...)

		switch expected := tt.expected.(type) {
		case []string:
			// Test successful split result
			arr, ok := result.(*Array)
			if !ok {
				t.Errorf("test %d: expected Array, got %T (%+v)", i, result, result)
				continue
			}
			if len(arr.Elements) != len(expected) {
				t.Errorf("test %d: expected %d elements, got %d", i, len(expected), len(arr.Elements))
				continue
			}
			for j, elem := range arr.Elements {
				str, ok := elem.(*String)
				if !ok {
					t.Errorf("test %d: element %d is not String, got %T", i, j, elem)
					continue
				}
				if str.Value != expected[j] {
					t.Errorf("test %d: element %d expected %q, got %q", i, j, expected[j], str.Value)
				}
			}
		case string:
			// Test error result
			errObj, ok := result.(*Error)
			if !ok {
				t.Errorf("test %d: expected Error, got %T (%+v)", i, result, result)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("test %d: expected error message %q, got %q", i, expected, errObj.Message)
			}
		}
	}
}
