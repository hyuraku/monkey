// JSON processing examples
// This demonstrates JSON stringification

puts("JSON Examples:");

// Create and stringify JSON
let person = {"name": "Alice", "age": 30, "city": "Tokyo"};

puts("Hash object:");
puts(person);

puts("JSON string:");
let json_output = json_stringify(person);
puts(json_output);

// Array to JSON
let numbers = [1, 2, 3, 4, 5];
puts("Array:");
puts(numbers);
puts("JSON array:");
puts(json_stringify(numbers));

// Nested structure
let hobbies = ["reading", "coding", "music"];
let user = {"name": "Bob", "hobbies": hobbies};
puts("Nested structure:");
puts(user);
puts("JSON output:");
puts(json_stringify(user));
