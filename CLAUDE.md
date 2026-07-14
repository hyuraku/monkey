# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go implementation of the Monkey programming language, featuring both a tree-walking interpreter and bytecode compiler/VM. The project demonstrates dual execution models for educational purposes based on Thorsten Ball's interpreter and compiler books.

## Development Commands

### Build Commands
```bash
# Build main interpreter executable
go build -o monkey

# Build REPL executable
go build -o monkey-repl ./cmd/monkey-repl

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
```

### Benchmark Commands
```bash
# Compare interpreter vs VM performance
./fibonacci -engine=vm     # Bytecode VM execution
./fibonacci -engine=eval   # Tree-walking interpreter
```

### Running Monkey Programs
```bash
# Execute a Monkey file
./monkey -e examples/hello.monkey

# Run the interactive REPL
./monkey-repl
```

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
