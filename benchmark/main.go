package main

import (
	"flag"
	"fmt"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/jit"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm', 'eval', or 'jit'")

var input = `
	let fibonacci = fn(x) {
		if (x == 0) {
			0
		} else {
			if (x == 1) {
				return 1;
			} else {
				fibonacci(x - 1) + fibonacci(x - 2);
			}
		}
	};
	fibonacci(35);
	`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object
	l := lexer.New(input)

	p := parser.New(l)
	program := p.ParseProgram()

	switch *engine {
	case "vm":
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}
		machine := vm.New(comp.Bytecode())
		start := time.Now()
		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}
		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	case "jit":
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		// JIT設定
		jitConfig := &jit.Config{
			Threshold:         1, // 低い閾値で即座にJIT化
			OptimizationLevel: 2,
			EnableProfiling:   true,
			MaxCodeCacheSize:  10 * 1024 * 1024,
		}

		// JIT統合VM作成
		originalVM := vm.New(comp.Bytecode())
		vmWithJIT := jit.NewVMWithJIT(originalVM, jitConfig)
		defer vmWithJIT.GetIntegration().Cleanup()

		start := time.Now()
		err = vmWithJIT.Run() // JIT統合VMでの実行
		if err != nil {
			fmt.Printf("jit vm error: %s", err)
			return
		}
		result = vmWithJIT.LastPoppedStackElem()
		duration = time.Since(start)
	default: // "eval"
		env := object.NewEnvironment()
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		*engine,
		result.Inspect(),
		duration)
}
