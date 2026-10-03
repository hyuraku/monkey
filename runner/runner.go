// Package runner executes a whole Monkey program (for example the contents of
// a .monkey file) on one of the two execution engines: the bytecode compiler +
// VM or the tree-walking evaluator.
package runner

import (
	"errors"
	"fmt"
	"strings"

	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
)

// Engine names accepted by the -engine flag (same values as benchmark/).
const (
	EngineVM   = "vm"
	EngineEval = "eval"
)

// IsValidEngine reports whether engine is one of the supported engine names.
func IsValidEngine(engine string) bool {
	return engine == EngineVM || engine == EngineEval
}

// Run parses input and executes it on the given engine. Output produced by
// the program itself (puts) is written by the builtins to standard output.
// A parse, compile or runtime error is returned as a non-nil error.
func Run(input, engine string) error {
	if !IsValidEngine(engine) {
		return fmt.Errorf("unknown engine %q (use %q or %q)", engine, EngineVM, EngineEval)
	}

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return errors.New("parser errors:\n\t" + strings.Join(p.Errors(), "\n\t"))
	}

	if engine == EngineEval {
		env := object.NewEnvironment()
		result := evaluator.Eval(program, env)
		if errObj, ok := result.(*object.Error); ok {
			return fmt.Errorf("evaluation failed: %s", errObj.Message)
		}
		return nil
	}

	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		return fmt.Errorf("executing bytecode failed: %w", err)
	}
	return nil
}
