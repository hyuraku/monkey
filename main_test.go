package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout runs fn and returns what it wrote to os.Stdout. The puts
// builtin writes to os.Stdout directly, so it cannot be redirected otherwise.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %s", err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	defer func() {
		os.Stdout = orig
	}()
	fn()
	_ = w.Close()
	return <-done
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.monkey")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %s", err)
	}
	return path
}

func TestRunFile(t *testing.T) {
	path := writeFile(t, `
let add = fn(a, b) {
	a + b
};
puts(add(1, 2));
`)

	for _, args := range [][]string{
		{path}, // default engine
		{"-engine=vm", path},
		{"-engine=eval", path},
	} {
		t.Run(strings.Join(args[:len(args)-1], " "), func(t *testing.T) {
			var stderr bytes.Buffer
			var code int
			out := captureStdout(t, func() {
				code = run(args, strings.NewReader(""), io.Discard, &stderr)
			})
			if code != 0 {
				t.Fatalf("exit code: got %d, want 0 (stderr: %q)", code, stderr.String())
			}
			if out != "3\n" {
				t.Fatalf("stdout: got %q, want %q", out, "3\n")
			}
		})
	}
}

func TestRunErrorExitCodes(t *testing.T) {
	syntaxError := writeFile(t, "let = 5;")
	runtimeError := writeFile(t, `puts("before"); 1 + true;`)
	missing := filepath.Join(t.TempDir(), "missing.monkey")

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{"missing file vm", []string{"-engine=vm", missing}, 1, "no such file"},
		{"missing file eval", []string{"-engine=eval", missing}, 1, "no such file"},
		{"syntax error vm", []string{"-engine=vm", syntaxError}, 1, "parser errors"},
		{"syntax error eval", []string{"-engine=eval", syntaxError}, 1, "parser errors"},
		{"runtime error vm", []string{"-engine=vm", runtimeError}, 1, "executing bytecode failed"},
		{"runtime error eval", []string{"-engine=eval", runtimeError}, 1, "evaluation failed"},
		{"invalid engine", []string{"-engine=jit", syntaxError}, 2, "Usage: monkey"},
		{"invalid engine without file", []string{"-engine=jit"}, 2, "Usage: monkey"},
		{"unknown flag", []string{"-e", syntaxError}, 2, "Usage: monkey"},
		{"too many arguments", []string{syntaxError, syntaxError}, 2, "too many arguments"},
		{"help", []string{"-h"}, 0, "Usage: monkey"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			var code int
			captureStdout(t, func() {
				code = run(tt.args, strings.NewReader(""), io.Discard, &stderr)
			})
			if code != tt.wantCode {
				t.Fatalf("exit code: got %d, want %d (stderr: %q)", code, tt.wantCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr: got %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

// examplesWithHashOutput print a hash with puts. Hash.Inspect iterates a Go
// map, so its key order is not deterministic and the outputs of the two
// engines cannot be compared byte for byte.
var examplesWithHashOutput = map[string]bool{
	"hashes.monkey": true,
	"json.monkey":   true,
}

func TestExamplesRunOnBothEngines(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("examples", "*.monkey"))
	if err != nil {
		t.Fatalf("Glob: %s", err)
	}
	if len(files) == 0 {
		t.Fatal("no examples found")
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			outputs := map[string]string{}
			for _, engine := range []string{"vm", "eval"} {
				var stderr bytes.Buffer
				var code int
				outputs[engine] = captureStdout(t, func() {
					code = run([]string{"-engine=" + engine, file}, strings.NewReader(""), io.Discard, &stderr)
				})
				if code != 0 {
					t.Fatalf("engine=%s: exit code %d, stderr: %s", engine, code, stderr.String())
				}
				if stderr.Len() != 0 {
					t.Fatalf("engine=%s: unexpected stderr: %s", engine, stderr.String())
				}
				if outputs[engine] == "" {
					t.Fatalf("engine=%s: no output", engine)
				}
			}

			if examplesWithHashOutput[filepath.Base(file)] {
				return
			}
			if outputs["vm"] != outputs["eval"] {
				t.Fatalf("outputs differ\nvm:\n%s\neval:\n%s", outputs["vm"], outputs["eval"])
			}
		})
	}
}

func TestREPLEngines(t *testing.T) {
	for _, engine := range []string{"vm", "eval"} {
		t.Run(engine, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("let x = 2;\nx * 21\nexit\n")
			var code int
			captureStdout(t, func() {
				code = run([]string{"-engine=" + engine}, stdin, &stdout, &stderr)
			})
			if code != 0 {
				t.Fatalf("exit code: got %d, want 0 (stderr: %q)", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), "42\n") {
				t.Fatalf("REPL output: got %q, want it to contain %q", stdout.String(), "42\n")
			}
		})
	}
}
