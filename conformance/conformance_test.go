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
