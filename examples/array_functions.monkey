// Array transformation examples
// This demonstrates working with arrays and functions

let numbers = [1, 2, 3, 4, 5];

puts("Original array:");
puts(numbers);

// Manual transformation
let double = fn(x) { x * 2 };
puts("Doubled values:");
puts(double(first(numbers)));
puts(double(first(rest(numbers))));
puts(double(first(rest(rest(numbers)))));

// Array concatenation and manipulation
let moreNumbers = [6, 7, 8];
puts("Another array:");
puts(moreNumbers);

puts("First elements:");
puts(first(numbers));
puts(first(moreNumbers));

puts("Last elements:");
puts(last(numbers));
puts(last(moreNumbers));

// Building new arrays
let squared = push(push(push([], 1), 4), 9);
puts("Squared array:");
puts(squared);
