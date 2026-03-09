// Array operations example
// This demonstrates array manipulation and built-in functions

let numbers = [1, 2, 3, 4, 5];
puts("Original array:");
puts(numbers);

puts("\nArray length:");
puts(len(numbers));

puts("\nFirst element:");
puts(first(numbers));

puts("\nLast element:");
puts(last(numbers));

puts("\nRest of array (without first):");
puts(rest(numbers));

puts("\nPush new element:");
let extended = push(numbers, 6);
puts(extended);

puts("\nPop last element:");
let popped = pop(extended);
puts(popped);
