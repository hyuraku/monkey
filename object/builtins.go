package object

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *Array:
				return NewInteger(int64(len(arg.Elements)))
			case *String:
				return NewInteger(int64(len(arg.Value)))
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
		},
	},
	{
		"puts",
		&Builtin{Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return nil
		},
		},
	},
	{
		"first",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `first` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[0]
			}

			return nil
		},
		},
	},
	{
		"last",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `last` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return nil
		},
		},
	},
	{
		"rest",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `rest` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &Array{Elements: newElements}
			}

			return nil
		},
		},
	},
	{
		"push",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `push` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)

			newElements := make([]Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &Array{Elements: newElements}
		},
		},
	},
	{
		"pop",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `pop` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `pop` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]Object, length-1, length-1)
				copy(newElements, arr.Elements[0:length-1])
				return &Array{Elements: newElements}
			}

			return nil
		},
		},
	},
	{
		"upper",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `upper` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("argument to `upper` must be STRING, got %s",
					args[0].Type())
			}

			str := args[0].(*String)
			return &String{Value: strings.ToUpper(str.Value)}
		},
		},
	},
	{
		"lower",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `lower` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("argument to `lower` must be STRING, got %s",
					args[0].Type())
			}

			str := args[0].(*String)
			return &String{Value: strings.ToLower(str.Value)}
		},
		},
	},
	{
		"split",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `split` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `split` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("first argument to `split` must be STRING, got %s",
					args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newError("second argument to `split` must be STRING, got %s",
					args[1].Type())
			}

			str := args[0].(*String)
			delimiter := args[1].(*String)

			if delimiter.Value == "" {
				return newError("delimiter cannot be empty")
			}

			parts := strings.Split(str.Value, delimiter.Value)
			elements := make([]Object, len(parts))
			for i, part := range parts {
				elements[i] = &String{Value: part}
			}

			return &Array{Elements: elements}
		},
		},
	},
	{
		"join",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `join` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `join` cannot be nil")
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `join` must be ARRAY, got %s",
					args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newError("second argument to `join` must be STRING, got %s",
					args[1].Type())
			}

			arr := args[0].(*Array)
			delimiter := args[1].(*String)

			// Convert all elements to strings using their Inspect() method
			stringElements := make([]string, len(arr.Elements))
			for i, elem := range arr.Elements {
				if elem == nil {
					stringElements[i] = ""
				} else {
					stringElements[i] = elem.Inspect()
				}
			}

			// Join the string elements with the delimiter
			result := strings.Join(stringElements, delimiter.Value)
			return &String{Value: result}
		},
		},
	},
	{
		"abs",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `abs` cannot be nil")
			}

			switch arg := args[0].(type) {
			case *Integer:
				value := arg.Value
				if value < 0 {
					value = -value
				}
				return NewInteger(value)
			case *Float:
				return &Float{Value: math.Abs(arg.Value)}
			default:
				return newError("argument to `abs` must be INTEGER or FLOAT, got %s",
					args[0].Type())
			}
		},
		},
	},
	{
		"min",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `min` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `min` cannot be nil")
			}

			// Check first argument type
			var val1 float64
			var isFloat1 bool
			switch arg := args[0].(type) {
			case *Integer:
				val1 = float64(arg.Value)
				isFloat1 = false
			case *Float:
				val1 = arg.Value
				isFloat1 = true
			default:
				return newError("first argument to `min` must be INTEGER or FLOAT, got %s",
					args[0].Type())
			}

			// Check second argument type
			var val2 float64
			var isFloat2 bool
			switch arg := args[1].(type) {
			case *Integer:
				val2 = float64(arg.Value)
				isFloat2 = false
			case *Float:
				val2 = arg.Value
				isFloat2 = true
			default:
				return newError("second argument to `min` must be INTEGER or FLOAT, got %s",
					args[1].Type())
			}

			// If either argument is float, result is float
			if isFloat1 || isFloat2 {
				return &Float{Value: math.Min(val1, val2)}
			}

			// Both are integers, return integer
			if val1 < val2 {
				return args[0]
			}
			return args[1]
		},
		},
	},
	{
		"max",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `max` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `max` cannot be nil")
			}

			// Check first argument type
			var val1 float64
			var isFloat1 bool
			switch arg := args[0].(type) {
			case *Integer:
				val1 = float64(arg.Value)
				isFloat1 = false
			case *Float:
				val1 = arg.Value
				isFloat1 = true
			default:
				return newError("first argument to `max` must be INTEGER or FLOAT, got %s",
					args[0].Type())
			}

			// Check second argument type
			var val2 float64
			var isFloat2 bool
			switch arg := args[1].(type) {
			case *Integer:
				val2 = float64(arg.Value)
				isFloat2 = false
			case *Float:
				val2 = arg.Value
				isFloat2 = true
			default:
				return newError("second argument to `max` must be INTEGER or FLOAT, got %s",
					args[1].Type())
			}

			// If either argument is float, result is float
			if isFloat1 || isFloat2 {
				return &Float{Value: math.Max(val1, val2)}
			}

			// Both are integers, return integer
			if val1 > val2 {
				return args[0]
			}
			return args[1]
		},
		},
	},
	{
		"sqrt",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `sqrt` cannot be nil")
			}

			var value float64
			switch arg := args[0].(type) {
			case *Integer:
				value = float64(arg.Value)
			case *Float:
				value = arg.Value
			default:
				return newError("argument to `sqrt` must be INTEGER or FLOAT, got %s",
					args[0].Type())
			}

			if value < 0 {
				return newError("sqrt of negative number is not supported")
			}

			return &Float{Value: math.Sqrt(value)}
		},
		},
	},
	{
		"regex",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `regex` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("argument to `regex` must be STRING, got %s",
					args[0].Type())
			}

			pattern := args[0].(*String).Value
			re, err := regexp.Compile(pattern)
			if err != nil {
				return newError("invalid regex pattern: %s", err.Error())
			}

			return &Regex{Pattern: pattern, Regexp: re}
		},
		},
	},
	{
		"match",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `match` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `match` cannot be nil")
			}
			if args[0].Type() != REGEX_OBJ {
				return newError("first argument to `match` must be REGEX, got %s",
					args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newError("second argument to `match` must be STRING, got %s",
					args[1].Type())
			}

			regex := args[0].(*Regex)
			text := args[1].(*String).Value

			matches := regex.Regexp.FindStringSubmatch(text)
			if matches == nil {
				return NULL
			}

			elements := make([]Object, len(matches))
			for i, match := range matches {
				elements[i] = &String{Value: match}
			}

			return &Array{Elements: elements}
		},
		},
	},
	{
		"replace",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `replace` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `replace` cannot be nil")
			}
			if args[2] == nil {
				return newError("third argument to `replace` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("first argument to `replace` must be STRING, got %s",
					args[0].Type())
			}
			if args[1].Type() != REGEX_OBJ {
				return newError("second argument to `replace` must be REGEX, got %s",
					args[1].Type())
			}
			if args[2].Type() != STRING_OBJ {
				return newError("third argument to `replace` must be STRING, got %s",
					args[2].Type())
			}

			text := args[0].(*String).Value
			regex := args[1].(*Regex)
			replacement := args[2].(*String).Value

			result := regex.Regexp.ReplaceAllString(text, replacement)
			return &String{Value: result}
		},
		},
	},
	{
		"regex_split",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `regex_split` cannot be nil")
			}
			if args[1] == nil {
				return newError("second argument to `regex_split` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("first argument to `regex_split` must be STRING, got %s",
					args[0].Type())
			}
			if args[1].Type() != REGEX_OBJ {
				return newError("second argument to `regex_split` must be REGEX, got %s",
					args[1].Type())
			}

			text := args[0].(*String).Value
			regex := args[1].(*Regex)

			parts := regex.Regexp.Split(text, -1)
			elements := make([]Object, len(parts))
			for i, part := range parts {
				elements[i] = &String{Value: part}
			}

			return &Array{Elements: elements}
		},
		},
	},
	{
		"json_parse",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0] == nil {
				return newError("argument to `json_parse` cannot be nil")
			}
			if args[0].Type() != STRING_OBJ {
				return newError("argument to `json_parse` must be STRING, got %s",
					args[0].Type())
			}

			jsonStr := args[0].(*String).Value

			// Parse JSON string
			var jsonValue interface{}
			err := json.Unmarshal([]byte(jsonStr), &jsonValue)
			if err != nil {
				return newError("invalid JSON: %s", err.Error())
			}

			// Convert Go interface{} to Monkey Object
			return convertGoValueToMonkeyObject(jsonValue)
		},
		},
	},
	{
		"json_stringify",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2",
					len(args))
			}
			if args[0] == nil {
				return newError("first argument to `json_stringify` cannot be nil")
			}

			// Convert Monkey Object to Go interface{}
			goValue, ok := convertMonkeyObjectToGoValue(args[0])
			if !ok {
				return newError("cannot convert object to JSON")
			}

			// Optional indent parameter
			var jsonBytes []byte
			var err error
			if len(args) == 2 {
				if args[1] == nil {
					return newError("second argument to `json_stringify` cannot be nil")
				}
				if args[1].Type() != STRING_OBJ {
					return newError("second argument to `json_stringify` must be STRING, got %s",
						args[1].Type())
				}
				indent := args[1].(*String).Value
				jsonBytes, err = json.MarshalIndent(goValue, "", indent)
			} else {
				jsonBytes, err = json.Marshal(goValue)
			}

			if err != nil {
				return newError("JSON stringify error: %s", err.Error())
			}

			return &String{Value: string(jsonBytes)}
		},
		},
	},
}

// convertGoValueToMonkeyObject converts Go interface{} to Monkey Object
func convertGoValueToMonkeyObject(value interface{}) Object {
	switch v := value.(type) {
	case nil:
		return NULL
	case bool:
		if v {
			return TRUE
		}
		return FALSE
	case float64:
		// JSON numbers are always float64, but check if it's actually an integer
		if v == float64(int64(v)) {
			return NewInteger(int64(v))
		}
		return &Float{Value: v}
	case string:
		return &String{Value: v}
	case []interface{}:
		elements := make([]Object, len(v))
		for i, elem := range v {
			elements[i] = convertGoValueToMonkeyObject(elem)
		}
		return &Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[HashKey]HashPair)
		for key, val := range v {
			keyObj := &String{Value: key}
			valueObj := convertGoValueToMonkeyObject(val)
			hashKey := keyObj.HashKey()
			pairs[hashKey] = HashPair{Key: keyObj, Value: valueObj}
		}
		return &Hash{Pairs: pairs}
	default:
		return NULL
	}
}

// convertMonkeyObjectToGoValue converts Monkey Object to Go interface{}
func convertMonkeyObjectToGoValue(obj Object) (interface{}, bool) {
	switch o := obj.(type) {
	case *Integer:
		return o.Value, true
	case *Float:
		return o.Value, true
	case *Boolean:
		return o.Value, true
	case *String:
		return o.Value, true
	case *Array:
		result := make([]interface{}, len(o.Elements))
		for i, elem := range o.Elements {
			val, ok := convertMonkeyObjectToGoValue(elem)
			if !ok {
				return nil, false
			}
			result[i] = val
		}
		return result, true
	case *Hash:
		result := make(map[string]interface{})
		for _, pair := range o.Pairs {
			if keyStr, ok := pair.Key.(*String); ok {
				val, ok := convertMonkeyObjectToGoValue(pair.Value)
				if !ok {
					return nil, false
				}
				result[keyStr.Value] = val
			}
		}
		return result, true
	case *Null:
		return nil, true
	default:
		return nil, false
	}
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, bi := range Builtins {
		if bi.Name == name {
			return bi.Builtin
		}
	}
	return nil
}
