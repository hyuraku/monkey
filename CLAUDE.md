# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of the Monkey programming language, featuring both a tree-walking interpreter and bytecode compiler/VM. The project demonstrates dual execution models for educational purposes based on Thorsten Ball's interpreter and compiler books.

## Development Commands

### Build Commands
```bash
# Build the monkey executable (REPL + file execution)
go build -o monkey

# Build benchmark tool
go build -o fibonacci ./benchmark
```

### Test Commands
```bash
# Run all tests
go test ./...

# Test specific components
go test ./lexer
go test ./parser
go test ./evaluator
go test ./compiler
go test ./vm
go test ./object
go test ./code

# Dual-execution conformance tests (evaluator vs VM)
go test ./conformance
```

### Benchmark Commands
```bash
# Compare interpreter vs VM performance
./fibonacci -engine=vm     # Bytecode VM execution
./fibonacci -engine=eval   # Tree-walking interpreter
```

### Running Monkey Programs
```bash
# Run the interactive REPL (type `exit` to quit)
./monkey
./monkey -engine=eval              # REPL on the tree-walking interpreter

# Run a file (multi-line programs are fine)
./monkey examples/hello.monkey               # bytecode VM (default)
./monkey -engine=eval examples/hello.monkey  # tree-walking interpreter
```

Usage is `monkey [-engine=vm|eval] [file]`. `-engine` defaults to `vm` (same flag name and values as `benchmark/`) and applies to both the REPL and file execution. The REPL still reads one line at a time; use file execution for multi-line programs. File execution prints only what the program outputs with `puts()` (the value of the last expression is not printed). Parse, compile and runtime errors go to stderr with exit code 1; an invalid `-engine` value or extra arguments print usage with exit code 2.

The CLI is split into `main.go` (flag parsing and exit codes, tested in `main_test.go`), `runner/` (runs a whole program on the selected engine) and `repl/` (`Start` for the VM, `StartEval` for the evaluator). `main_test.go` runs every `examples/*.monkey` on both engines.

Note: the VM does not stop on errors returned by builtins (e.g. `first(1)` becomes an error value and execution continues), while the evaluator stops and file execution exits with code 1.

## Architecture Overview

The codebase implements two execution paths:

**Interpreter Path**: Source → Lexer → Parser → AST → Evaluator
**Compiler Path**: Source → Lexer → Parser → AST → Compiler → Bytecode → VM

### Key Components

- **lexer/**: Tokenizes source code into tokens
- **parser/**: Recursive descent parser building AST from tokens
- **ast/**: AST node definitions and interfaces
- **evaluator/**: Tree-walking interpreter that directly evaluates AST
- **compiler/**: Compiles AST to bytecode instructions
- **vm/**: Stack-based virtual machine executing bytecode
- **code/**: Bytecode instruction set definitions
- **object/**: Runtime object system and built-in functions
- **token/**: Token type definitions and keyword mappings
- **runner/**: Runs a whole program (file execution mode) on the VM or the evaluator
- **repl/**: Line-by-line REPL on either engine

### Object System

The runtime supports:
- Primitives: integers, floats, booleans, strings
- Collections: arrays, hashes
- Functions: first-class with closures
- Built-ins: `len()`, `first()`, `last()`, `rest()`, `push()`, `pop()`, `puts()`
- String processing: `upper()`, `lower()`, `split()`, `join()`
- Math functions: `abs()`, `min()`, `max()`, `sqrt()`
- Regular expressions: `regex()`, `match()`, `replace()`, `regex_split()`
- JSON processing: `json_parse()`, `json_stringify()`

When adding a new built-in function, define it in `object/builtins.go` only; both the evaluator and the compiler/VM register builtins from `object.Builtins` automatically (verified by the dual-execution tests in `conformance/`).

### Symbol Management

- **Environment** (interpreter): Lexical scoping with environment chaining
- **Symbol Table** (compiler): Compile-time symbol resolution for bytecode generation

## Testing Strategy

Each major component has comprehensive unit tests. The test suite covers lexical analysis, parsing, AST evaluation, bytecode compilation, and VM execution. All tests should pass before committing changes.

## Language Features

Monkey supports variable bindings, functions, closures, arrays, hashes, if expressions, and arithmetic/comparison operators. Recent additions include float support and the `pop()` built-in function.
