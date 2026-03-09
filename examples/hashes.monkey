// Hash (dictionary) operations
// This demonstrates hash creation and access

let person = {"name": "Alice", "age": 30, "city": "Tokyo"};

puts("Person hash:");
puts(person);

puts("Accessing hash values:");
puts(person["name"]);
puts(person["age"]);
puts(person["city"]);

// Nested hashes
let employees = {"engineer": 50, "designer": 20, "manager": 10};
let company = {"name": "Tech Corp", "employees": employees};

puts("Nested hash access:");
puts(company["name"]);
puts("Number of engineers:");
puts(company["employees"]["engineer"]);
