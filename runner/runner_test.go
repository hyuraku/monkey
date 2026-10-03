package runner

import (
	"io"
	"os"
	"strings"
	"testing"
)

var engines = []string{EngineVM, EngineEval}

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

func TestIsValidEngine(t *testing.T) {
	tests := []struct {
		engine string
		want   bool
	}{
		{"vm", true},
		{"eval", true},
		{"", false},
		{"VM", false},
		{"interpreter", false},
	}

	for _, tt := range tests {
		if got := IsValidEngine(tt.engine); got != tt.want {
			t.Errorf("IsValidEngine(%q) = %t, want %t", tt.engine, got, tt.want)
		}
	}
}

func TestRunMultiLineProgram(t *testing.T) {
	input := `
let fibonacci = fn(x) {
	if (x < 2) {
		x
	} else {
		fibonacci(x - 1) + fibonacci(x - 2)
	}
};
puts(fibonacci(10));
puts("done");
`
	want := "55\ndone\n"

	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			var err error
			got := captureStdout(t, func() {
				err = Run(input, engine)
			})
			if err != nil {
				t.Fatalf("Run returned error: %s", err)
			}
			if got != want {
				t.Fatalf("output: got %q, want %q", got, want)
			}
		})
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		engine  string
		wantErr string
	}{
		{"parse error vm", "let = 5;", EngineVM, "parser errors"},
		{"parse error eval", "let = 5;", EngineEval, "parser errors"},
		{"compile error vm", "puts(x);", EngineVM, "compilation failed: undefined variable x"},
		{"undefined identifier eval", "puts(x);", EngineEval, "evaluation failed: identifier not found: x"},
		{"runtime error vm", "1 + true;", EngineVM, "executing bytecode failed"},
		{"runtime error eval", "1 + true;", EngineEval, "evaluation failed: type mismatch"},
		{"unknown engine", "1;", "jit", "unknown engine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.input, tt.engine)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error: got %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestRunStopsAtRuntimeError(t *testing.T) {
	input := `puts("before"); 1 + true; puts("after");`

	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			var err error
			got := captureStdout(t, func() {
				err = Run(input, engine)
			})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if got != "before\n" {
				t.Fatalf("output: got %q, want %q", got, "before\n")
			}
		})
	}
}

func TestRunStopsAtBuiltinError(t *testing.T) {
	input := `puts("before"); first(1); puts("after");`
	wantErr := map[string]string{
		EngineVM:   "executing bytecode failed: argument to `first` must be ARRAY, got INTEGER",
		EngineEval: "evaluation failed: argument to `first` must be ARRAY, got INTEGER",
	}

	for _, engine := range engines {
		t.Run(engine, func(t *testing.T) {
			var err error
			got := captureStdout(t, func() {
				err = Run(input, engine)
			})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != wantErr[engine] {
				t.Fatalf("error: got %q, want %q", err.Error(), wantErr[engine])
			}
			if got != "before\n" {
				t.Fatalf("output: got %q, want %q", got, "before\n")
			}
		})
	}
}
