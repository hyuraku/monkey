# Monkey Programming Language

This is an implementation of the interpreter and compiler for the Monkey programming language. This project is based on the language specifications introduced in the books "Writing An Interpreter In Go" and "Writing A Compiler In Go".

## Features

- Simple yet powerful syntax
- First-class functions
- Closures
- Basic data types including integers, strings, arrays, and hashes
- Implementation of all steps from lexical analysis to execution (Lexer → Parser → AST → Evaluator)
- Bytecode compiler and virtual machine implementation

## How to Run

```bash
# Build the project
go build -o monkey

# Start the REPL
./monkey
```

Once the REPL starts, you can enter expressions in the Monkey language:

```monkey
>> let x = 10;
>> let y = 5;
>> x + y;
15
>> let add = fn(a, b) { a + b };
>> add(x, y);
15
>> let arr = [1, 2, 3];
>> arr[1];
2
```

## Project Structure

```
.
├── ast/                  # Abstract Syntax Tree definitions
├── code/                 # Bytecode instruction set definitions
├── compiler/             # Compiler implementation
├── evaluator/            # Interpreter implementation
├── lexer/                # Lexical analyzer
├── main.go               # Program entry point
├── object/               # Object system
├── parser/               # Parser
├── repl/                 # REPL (Read-Eval-Print Loop)
├── token/                # Token definitions
└── vm/                   # Virtual machine implementation
```

## Implemented Features

### Data Types

- Integers: `5`, `10`, `-5`
- Booleans: `true`, `false`
- Strings: `"hello world"`
- Arrays: `[1, 2, 3]`
- Hashes (associative arrays): `{"name": "Monkey", "age": 5}`
- Functions: `fn(x, y) { x + y }`

### Operators

- Arithmetic operators: `+`, `-`, `*`, `/`
- Comparison operators: `==`, `!=`, `<`, `>`
- Logical operators: `!` (negation)

### Control Flow

- If expressions: `if (x > y) { x } else { y }`
- Return statements: `return x + y;`

### Bindings

- Variables: `let x = 5;`
- Functions: `let add = fn(x, y) { x + y };`

### Built-in Functions

- `len()`: Returns the length of arrays or strings
- `first()`: Returns the first element of an array
- `last()`: Returns the last element of an array
- `rest()`: Returns the rest of the array excluding the first element
- `push()`: Adds an element to an array
- `puts()`: Outputs a value to standard output

## Testing

Each package in the project includes tests. To run the tests:

```bash
# To test a specific package
go test ./lexer
```

## References

- "Writing An Interpreter In Go" by Thorsten Ball
- "Writing A Compiler In Go" by Thorsten Ball Monkey プログラミング言語
