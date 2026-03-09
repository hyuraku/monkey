// Functions and first-class functions
// This demonstrates function definitions and usage

let add = fn(a, b) { a + b };
let multiply = fn(a, b) { a * b };
let subtract = fn(a, b) { a - b };
let divide = fn(a, b) { a / b };

puts("Basic arithmetic functions:");
puts(add(10, 5));
puts(subtract(10, 5));
puts(multiply(10, 5));
puts(divide(10, 5));

// Function composition
let double = fn(x) { x * 2 };
let square = fn(x) { x * x };

puts("Function application:");
puts(double(5));
puts(square(5));
puts(double(square(5)));
