# Monkey Language Examples

This directory contains example programs written in the Monkey programming language, demonstrating various features and capabilities.

## Running Examples

First, make sure you have built the Monkey interpreter:

```bash
go build -o monkey
```

Then you can run any example file using:

```bash
./monkey examples/filename.monkey
```

Add `-engine=eval` to run it on the tree-walking interpreter instead of the bytecode VM (the default):

```bash
./monkey -engine=eval examples/filename.monkey
```

Or run the examples from the examples directory:

```bash
cd examples
../monkey filename.monkey
```

## Available Examples

### 1. Hello World
**File**: [hello.monkey](hello.monkey)

Basic string output demonstration.

```bash
../monkey hello.monkey
```

### 2. Recursion
**File**: [recursion.monkey](recursion.monkey)

Recursive function examples demonstrating countdown and sum calculations.

```bash
../monkey recursion.monkey
```

### 3. Array Operations
**File**: [arrays.monkey](arrays.monkey)

Demonstrates array manipulation with built-in functions: `len()`, `first()`, `last()`, `rest()`, `push()`, `pop()`.

```bash
../monkey arrays.monkey
```

### 4. Functions
**File**: [functions.monkey](functions.monkey)

First-class functions and function composition examples.

```bash
../monkey functions.monkey
```

### 5. Hash Operations
**File**: [hashes.monkey](hashes.monkey)

Hash (dictionary) creation and access, including nested hashes.

```bash
../monkey hashes.monkey
```

### 6. Array Functions
**File**: [array_functions.monkey](array_functions.monkey)

Working with arrays and functions for data transformation.

```bash
../monkey array_functions.monkey
```

### 7. String Operations
**File**: [string_operations.monkey](string_operations.monkey)

String manipulation using built-in functions: `upper()`, `lower()`, `split()`, `join()`.

```bash
../monkey string_operations.monkey
```

### 8. Math Operations
**File**: [math.monkey](math.monkey)

Mathematical functions and float support: `abs()`, `min()`, `max()`, `sqrt()`.

```bash
../monkey math.monkey
```

### 9. Regular Expressions
**File**: [regex.monkey](regex.monkey)

Pattern matching and text manipulation using `regex()`, `match()`, `replace()`, `regex_split()`.

```bash
../monkey regex.monkey
```

### 10. JSON Processing
**File**: [json.monkey](json.monkey)

JSON stringification with `json_stringify()` for converting arrays and hashes to JSON format.

```bash
../monkey json.monkey
```

## Running All Examples

You can run all examples at once using a simple shell loop:

```bash
# From the project root
for file in examples/*.monkey; do
    echo "==== Running $file ===="
    ./monkey "$file"
    echo ""
done
```

Or from the examples directory:

```bash
# From the examples directory
for file in *.monkey; do
    echo "==== Running $file ===="
    ../monkey "$file"
    echo ""
done
```

## Interactive REPL

You can also try these code snippets interactively in the REPL by running the binary without a file:

```bash
./monkey
```

Type or paste Monkey code into the prompt. The REPL reads one line at a time, so multi-line programs are easier to run as files.

## Language Features Demonstrated

- **Variables and Bindings**: `let` keyword for variable declarations
- **Functions**: First-class functions and function composition
- **Data Types**: integers, floats, booleans, strings, arrays, hashes
- **Control Flow**: `if/else` expressions
- **Operators**: arithmetic (`+`, `-`, `*`, `/`), comparison (`==`, `!=`, `<`, `>`)
- **Built-in Functions**: 20+ built-in functions for common operations
- **Recursion**: Full support for recursive function calls
- **String Operations**: `upper()`, `lower()`, `split()`, `join()`
- **Math Functions**: `abs()`, `min()`, `max()`, `sqrt()`
- **Regular Expressions**: `regex()`, `match()`, `replace()`, `regex_split()`
- **JSON**: `json_stringify()` for serialization

## Creating Your Own Examples

Feel free to create your own `.monkey` files in this directory and run them using the same commands. The Monkey language is designed to be simple and expressive, making it easy to experiment with different programming concepts.

## Benchmarking

For performance comparisons between the VM and interpreter, see the [benchmark](../benchmark) directory which includes the Fibonacci benchmark tool.
