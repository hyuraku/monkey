package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"monkey/repl"
	"monkey/runner"
	"os"
	"os/user"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run parses the command line, then starts the REPL (no file argument) or
// executes the given file. It returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("monkey", flag.ContinueOnError)
	flags.SetOutput(stderr)
	engine := flags.String("engine", runner.EngineVM, "use 'vm' or 'eval'")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage: monkey [-engine=vm|eval] [file]\n\n")
		_, _ = fmt.Fprintf(stderr, "Starts the REPL when no file is given, otherwise runs the file.\n\n")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if !runner.IsValidEngine(*engine) {
		_, _ = fmt.Fprintf(stderr, "invalid value %q for flag -engine: use 'vm' or 'eval'\n", *engine)
		flags.Usage()
		return 2
	}

	switch flags.NArg() {
	case 0:
		return startREPL(stdin, stdout, *engine)
	case 1:
		return runFile(flags.Arg(0), *engine, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "too many arguments: %v\n", flags.Args())
		flags.Usage()
		return 2
	}
}

func startREPL(stdin io.Reader, stdout io.Writer, engine string) int {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprintf(stdout, "Hello %s! This is the Monkey programming language!\n", user.Username)
	_, _ = fmt.Fprintf(stdout, "Feel free to type in commands\n")
	if engine == runner.EngineEval {
		repl.StartEval(stdin, stdout)
	} else {
		repl.Start(stdin, stdout)
	}
	return 0
}

func runFile(path, engine string, stderr io.Writer) int {
	input, err := os.ReadFile(path)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "monkey: %s\n", err)
		return 1
	}
	if err := runner.Run(string(input), engine); err != nil {
		_, _ = fmt.Fprintf(stderr, "monkey: %s: %s\n", path, err)
		return 1
	}
	return 0
}
