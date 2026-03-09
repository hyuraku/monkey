// Recursive function example
// This demonstrates simple recursion

let countdown = fn(n) { if (n < 1) { n } else { countdown(n - 1) } };

puts("Countdown examples:");
puts(countdown(0));
puts(countdown(1));
puts(countdown(5));
puts(countdown(10));

let sumTo = fn(n) { if (n < 1) { 0 } else { n + sumTo(n - 1) } };

puts("Sum examples:");
puts(sumTo(0));
puts(sumTo(1));
puts(sumTo(5));
puts(sumTo(10));
