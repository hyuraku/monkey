// Regular expression examples
// This demonstrates regex pattern matching and manipulation

puts("Regular Expression Examples:");

let pattern = regex("[0-9]+");
let text = "I have 42 apples and 7 oranges";

puts("Original text:");
puts(text);

puts("Match numbers:");
puts(match(pattern, text));

puts("Replace numbers with X:");
let replaced = replace(text, pattern, "X");
puts(replaced);

puts("Split by numbers:");
let parts = regex_split(text, pattern);
puts(parts);

// Simple pattern matching
puts("Pattern matching:");
let word_pattern = regex("world");
let text1 = "hello world";
let text2 = "helloearth";

puts(text1);
puts(match(word_pattern, text1));

puts(text2);
puts(match(word_pattern, text2));
