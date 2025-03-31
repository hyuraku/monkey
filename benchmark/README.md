# Monkey Language Benchmarks

This directory contains benchmarks for measuring the performance of the Monkey language interpreter and compiler.

## Overview

There are two approaches to implementing the Monkey language:
1. Interpreter approach - directly evaluates the AST
2. Compiler/VM approach - compiles the AST into bytecode and executes it in a VM

We've prepared benchmarks to measure the performance differences between these implementations.

## Running the Benchmarks

To run the benchmarks, use the following commands:

```bash
cd ../
go build -o fibonacci ./benchmark
./fibonacci -engine=vm
./fibonacci -engine=eval
```
