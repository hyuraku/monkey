# Monkey Programming Language

This is an implementation of the interpreter and compiler for the Monkey programming language. This project is based on the language specifications introduced in the books "Writing An Interpreter In Go" and "Writing A Compiler In Go".

## Features

- Simple yet powerful syntax
- First-class functions and closures
- Rich data types: integers, floats, booleans, strings, arrays, hashes, and regular expressions
- Comprehensive built-in functions for string processing, math operations, and pattern matching
- Dual execution models: tree-walking interpreter and bytecode compiler with virtual machine
- Memory-optimized object system with singleton instances and integer caching

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
>> x += 3;  // Assignment operators
13
>> y *= 2;
10
>> let add = fn(a, b) { a + b };
>> add(x, y);
23
>> let arr = [1, 2, 3];
>> arr[1];
2
>> // String processing and regex
>> let text = "Hello 123 World";
>> let numRegex = regex("\\d+");
>> match(numRegex, text);
[123]
>> replace(text, numRegex, "XXX");
Hello XXX World
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

- **Integers**: `5`, `10`, `-5`
- **Floats**: `3.14`, `-5.2`
- **Booleans**: `true`, `false`
- **Strings**: `"hello world"`
- **Arrays**: `[1, 2, 3]`
- **Hashes**: `{"name": "Monkey", "age": 5}`
- **Functions**: `fn(x, y) { x + y }`
- **Regular Expressions**: Created with `regex("pattern")`

### Operators

- Arithmetic operators: `+`, `-`, `*`, `/`
- Assignment operators: `+=`, `-=`, `*=`, `/=`
- Comparison operators: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical operators: `!` (negation), `&&`, `||`

### Control Flow

- If expressions: `if (x > y) { x } else { y }`
- Return statements: `return x + y;`

### Bindings

- Variables: `let x = 5;`
- Assignment operators: `x += 10;`, `y -= 5;`, `z *= 2;`, `w /= 3;`
- Functions: `let add = fn(x, y) { x + y };`

### Built-in Functions

#### Array and String Operations
- `len(obj)`: Returns the length of arrays or strings
- `first(array)`: Returns the first element of an array
- `last(array)`: Returns the last element of an array
- `rest(array)`: Returns the rest of the array excluding the first element
- `push(array, element)`: Adds an element to an array
- `pop(array)`: Removes and returns the last element of an array

#### String Processing
- `upper(string)`: Converts string to uppercase
- `lower(string)`: Converts string to lowercase
- `split(string, delimiter)`: Splits string by delimiter into array
- `join(array, delimiter)`: Joins array elements into string with delimiter

#### Regular Expression Operations
- `regex(pattern)`: Creates a regular expression object from pattern string
- `match(regex, text)`: Returns array of matches or null if no match
- `replace(text, regex, replacement)`: Replaces all matches with replacement string
- `regex_split(text, regex)`: Splits text using regular expression pattern

#### Math Functions
- `abs(number)`: Returns absolute value of number
- `min(a, b)`: Returns the smaller of two numbers
- `max(a, b)`: Returns the larger of two numbers
- `sqrt(number)`: Returns square root of number (as float)

#### JSON Processing
- `json_parse(json_string)`: Parses a JSON string and returns the corresponding Monkey object
- `json_stringify(object [, indent])`: Converts a Monkey object to JSON string with optional indentation

#### I/O
- `puts(...)`: Outputs values to standard output

## Usage Examples

### Pattern Matching and Text Processing

```monkey
// Email extraction
let emailRegex = regex("\\w+@\\w+\\.\\w+");
let text = "Contact us at support@example.com for help";
let result = match(emailRegex, text);
puts(result[0]); // support@example.com

// Text replacement
let phoneText = "Call 555-1234 or 555-5678";
let phoneRegex = regex("\\d{3}-\\d{4}");
let masked = replace(phoneText, phoneRegex, "XXX-XXXX");
puts(masked); // Call XXX-XXXX or XXX-XXXX

// Text splitting
let csvData = "apple,banana,cherry,date";
let items = split(csvData, ",");
puts(items); // [apple, banana, cherry, date]

// Regex-based splitting
let whitespaceText = "word1   word2     word3";
let wsRegex = regex("\\s+");
let words = regex_split(whitespaceText, wsRegex);
puts(words); // [word1, word2, word3]
```

### Mathematical Operations

```monkey
// Basic math functions
puts(abs(-42));        // 42
puts(min(10, 5));      // 5
puts(max(10, 5));      // 10
puts(sqrt(16));        // 4.000000

// Working with arrays
let numbers = [3, 1, 4, 1, 5, 9];
puts(len(numbers));    // 6
puts(first(numbers));  // 3
puts(last(numbers));   // 9

let more_numbers = push(numbers, 2);
let fewer_numbers = pop(numbers);
```

### String Manipulation

```monkey
let message = "Hello World";
puts(upper(message));  // HELLO WORLD
puts(lower(message));  // hello world

let words = ["Hello", "beautiful", "world"];
let sentence = join(words, " ");
puts(sentence);        // Hello beautiful world

let parts = split(sentence, " ");
puts(len(parts));      // 3
```

### JSON Processing

```monkey
// Parsing JSON strings
let jsonStr = `{"name": "Alice", "age": 30, "active": true}`;
let data = json_parse(jsonStr);
puts(data["name"]);     // Alice
puts(data["age"]);      // 30

// Working with JSON arrays
let arrayJson = `[1, 2, 3, {"nested": "value"}]`;
let array = json_parse(arrayJson);
puts(array[0]);         // 1
puts(array[3]["nested"]); // value

// Converting objects to JSON
let person = {"name": "Bob", "age": 25, "hobbies": ["reading", "coding"]};
let compact = json_stringify(person);
puts(compact);          // {"age":25,"hobbies":["reading","coding"],"name":"Bob"}

// Pretty-printing with indentation
let pretty = json_stringify(person, "  ");
puts(pretty);
// {
//   "age": 25,
//   "hobbies": [
//     "reading",
//     "coding"
//   ],
//   "name": "Bob"
// }

// Working with complex nested structures
let config = {
  "database": {
    "host": "localhost",
    "port": 5432,
    "credentials": {
      "username": "admin",
      "password": "secret"
    }
  },
  "features": ["auth", "logging", "metrics"]
};

let configJson = json_stringify(config, "  ");
let parsed = json_parse(configJson);
puts(parsed["database"]["host"]);  // localhost
```

## Performance Comparison

This implementation includes both a tree-walking interpreter and a bytecode compiler with virtual machine. You can compare their performance using the included benchmark:

```bash
# Build the benchmark tool
go build -o fibonacci ./benchmark

# Run with interpreter (tree-walking)
./fibonacci -engine=eval

# Run with virtual machine (bytecode)
./fibonacci -engine=vm
```

The virtual machine typically shows significant performance improvements over the interpreter for computational tasks.

## Testing

Each package in the project includes comprehensive tests. To run the tests:

```bash
# Run all tests
go test ./...

# Test a specific package
go test ./lexer
go test ./parser
go test ./object
go test ./evaluator
go test ./compiler
go test ./vm
```

## References

- "Writing An Interpreter In Go" by Thorsten Ball
- "Writing A Compiler In Go" by Thorsten Ball
