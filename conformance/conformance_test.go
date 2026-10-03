// Package conformance provides dual-execution tests that run the same Monkey
// program on both execution engines (the tree-walking evaluator and the
// bytecode compiler + VM) and verify they produce the same result.
package conformance

import (
	"testing"

	"monkey/ast"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
)

func parse(t *testing.T, input string) *ast.Program {
	t.Helper()
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors for %q: %v", input, p.Errors())
	}
	return program
}

func runEval(t *testing.T, input string) object.Object {
	t.Helper()
	program := parse(t, input)
	env := object.NewEnvironment()
	return evaluator.Eval(program, env)
}

func runVM(t *testing.T, input string) object.Object {
	t.Helper()
	program := parse(t, input)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compile error for %q: %s", input, err)
	}
	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		t.Fatalf("vm error for %q: %s", input, err)
	}
	return machine.LastPoppedStackElem()
}

// builtinCases covers one normal-path case for each of the 21 builtin
// functions defined in object/builtins.go. Every case must produce the
// same result on both execution engines.
var builtinCases = []struct {
	name  string
	input string
	want  string // expected Inspect() output on both paths
}{
	{"len", `len("hello")`, "5"},
	{"puts", `puts("conformance")`, "null"},
	{"first", `first([1, 2, 3])`, "1"},
	{"last", `last([1, 2, 3])`, "3"},
	{"rest", `rest([1, 2, 3])`, "[2, 3]"},
	{"push", `push([1, 2], 3)`, "[1, 2, 3]"},
	{"pop", `pop([1, 2, 3])`, "[1, 2]"},
	{"upper", `upper("monkey")`, "MONKEY"},
	{"lower", `lower("MONKEY")`, "monkey"},
	{"split", `split("a,b,c", ",")`, "[a, b, c]"},
	{"join", `join(["a", "b", "c"], "-")`, "a-b-c"},
	{"abs", `abs(-5)`, "5"},
	{"min", `min(3, 1)`, "1"},
	{"max", `max(3, 1)`, "3"},
	{"sqrt", `sqrt(4)`, "2.000000"},
	{"regex", `regex("a+")`, "/a+/"},
	{"match", `match(regex("a+"), "caat")`, "[aa]"},
	{"replace", `replace("monkey", regex("o"), "0")`, "m0nkey"},
	{"regex_split", `regex_split("a,b", regex(","))`, "[a, b]"},
	{"json_parse", `json_parse("[1, 2, 3]")`, "[1, 2, 3]"},
	{"json_stringify", `json_stringify([1, 2, 3])`, "[1,2,3]"},
}

func TestBuiltinsDualExecution(t *testing.T) {
	for _, tc := range builtinCases {
		t.Run(tc.name, func(t *testing.T) {
			vmResult := runVM(t, tc.input)
			if vmResult == nil {
				t.Fatalf("vm returned nil for %q", tc.input)
			}
			if errObj, ok := vmResult.(*object.Error); ok {
				t.Fatalf("vm returned error for %q: %s", tc.input, errObj.Message)
			}
			if got := vmResult.Inspect(); got != tc.want {
				t.Fatalf("vm result for %q: got %q, want %q", tc.input, got, tc.want)
			}

			evalResult := runEval(t, tc.input)
			if evalResult == nil {
				t.Fatalf("evaluator returned nil for %q", tc.input)
			}

			if errObj, ok := evalResult.(*object.Error); ok {
				t.Fatalf("evaluator returned error for %q: %s", tc.input, errObj.Message)
			}
			if got := evalResult.Inspect(); got != tc.want {
				t.Fatalf("evaluator result for %q: got %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// builtinErrorCases are programs where a builtin returns an error. Both
// engines must stop there with the same message: the evaluator returns the
// *object.Error and the VM returns it as a Run error.
var builtinErrorCases = []struct {
	name    string
	input   string
	wantErr string
}{
	{"first non-array", `first(1); 99`, "argument to `first` must be ARRAY, got INTEGER"},
	{"len wrong arguments", `len("a", "b"); 99`, "wrong number of arguments. got=2, want=1"},
	{"error in let", `let x = last(1); 99`, "argument to `last` must be ARRAY, got INTEGER"},
	{"error inside function", `let f = fn() { push(1, 2); 1 }; f() + 1`, "argument to `push` must be ARRAY, got INTEGER"},
	{"json_parse invalid", `json_parse("{"); 99`, "invalid JSON: unexpected end of JSON input"},
}

func TestBuiltinErrorsDualExecution(t *testing.T) {
	for _, tc := range builtinErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			program := parse(t, tc.input)
			comp := compiler.New()
			if err := comp.Compile(program); err != nil {
				t.Fatalf("compile error for %q: %s", tc.input, err)
			}
			machine := vm.New(comp.Bytecode())
			err := machine.Run()
			if err == nil {
				t.Fatalf("vm: expected error for %q, got nil (last popped: %v)", tc.input, machine.LastPoppedStackElem())
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("vm error for %q: got %q, want %q", tc.input, err.Error(), tc.wantErr)
			}

			evalResult := runEval(t, tc.input)
			errObj, ok := evalResult.(*object.Error)
			if !ok {
				t.Fatalf("evaluator: expected error for %q, got %T (%v)", tc.input, evalResult, evalResult)
			}
			if errObj.Message != tc.wantErr {
				t.Fatalf("evaluator error for %q: got %q, want %q", tc.input, errObj.Message, tc.wantErr)
			}
		})
	}
}

// hashInspectCases check that a hash prints its pairs in the same, fixed
// order on both engines (Hash.Inspect sorts the keys).
var hashInspectCases = []struct {
	name  string
	input string
	want  string
}{
	{"string keys", `{"name": "Alice", "age": 30, "city": "Tokyo"}`, "{age: 30, city: Tokyo, name: Alice}"},
	{"integer keys", `{10: "ten", 2: "two", -1: "minus one"}`, "{-1: minus one, 2: two, 10: ten}"},
	{"mixed keys", `{"b": 1, 2: 2, true: 3, "a": 4}`, "{true: 3, 2: 2, a: 4, b: 1}"},
	{"nested", `{"z": {"y": 1, "x": 2}, "a": [1, 2]}`, "{a: [1, 2], z: {x: 2, y: 1}}"},
}

func TestHashInspectDualExecution(t *testing.T) {
	for _, tc := range hashInspectCases {
		t.Run(tc.name, func(t *testing.T) {
			// Go map iteration order is random, so run each case several times.
			for i := 0; i < 10; i++ {
				if got := runVM(t, tc.input).Inspect(); got != tc.want {
					t.Fatalf("vm result for %q: got %q, want %q", tc.input, got, tc.want)
				}
				if got := runEval(t, tc.input).Inspect(); got != tc.want {
					t.Fatalf("evaluator result for %q: got %q, want %q", tc.input, got, tc.want)
				}
			}
		})
	}
}

// escapeCases cover string escape sequences (handled by the lexer) and the
// builtins that receive such strings: regex patterns and JSON.
var escapeCases = []struct {
	name  string
	input string
	want  string
}{
	{"newline", `"a\nb"`, "a\nb"},
	{"tab and carriage return", `"a\tb\rc"`, "a\tb\rc"},
	{"backslash", `len("a\\b")`, "3"},
	{"quote", `"say \"hi\""`, `say "hi"`},
	{"split on newline", `split("a\nb\nc", "\n")`, "[a, b, c]"},
	{"regex escaped backslash", `match(regex("\\d+"), "Hello 123 World")`, "[123]"},
	{"regex unknown escape", `match(regex("\d+"), "Hello 123 World")`, "[123]"},
	{"regex replace", `replace("Hello 123 World", regex("\\d+"), "XXX")`, "Hello XXX World"},
	{"regex email", `match(regex("\\w+@\\w+\\.\\w+"), "Contact support@example.com")`, "[support@example.com]"},
	{"regex_split whitespace", `regex_split("a  b\tc", regex("\\s+"))`, "[a, b, c]"},
	{"json_stringify escapes control characters", `json_stringify("a\nb\t\"c\"\\")`, `"a\nb\t\"c\"\\"`},
	{"json_parse escaped quotes", `json_parse("{\"b\": 1, \"a\": \"x\"}")`, "{a: x, b: 1}"},
	{"json_parse json escapes", `json_parse("\"line1\\nline2\"")`, "line1\nline2"},
	{"json round trip", `json_parse(json_stringify("tab\there\nnew \"q\" \\"))`, "tab\there\nnew \"q\" \\"},
}

func TestStringEscapesDualExecution(t *testing.T) {
	for _, tc := range escapeCases {
		t.Run(tc.name, func(t *testing.T) {
			vmResult := runVM(t, tc.input)
			if got := vmResult.Inspect(); got != tc.want {
				t.Fatalf("vm result for %q: got %q, want %q", tc.input, got, tc.want)
			}

			evalResult := runEval(t, tc.input)
			if errObj, ok := evalResult.(*object.Error); ok {
				t.Fatalf("evaluator returned error for %q: %s", tc.input, errObj.Message)
			}
			if got := evalResult.Inspect(); got != tc.want {
				t.Fatalf("evaluator result for %q: got %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
