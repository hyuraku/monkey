// String manipulation examples
// This demonstrates string built-in functions

let text = "Hello, Monkey Language!";

puts("Original text:");
puts(text);

puts("Uppercase:");
puts(upper(text));

puts("Lowercase:");
puts(lower(text));

puts("Split by space:");
let words = split(text, " ");
puts(words);

puts("Join with hyphen:");
puts(join(words, "-"));

// String operations
let sentence = "The quick brown fox jumps";
puts("Original sentence:");
puts(sentence);

puts("Split into words:");
let wordArray = split(sentence, " ");
puts(wordArray);

puts("First word:");
puts(first(wordArray));

puts("Last word:");
puts(last(wordArray));
