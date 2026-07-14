package parser

import (
	"strings"
	"testing"
)

func TestFormatDiagnostic(t *testing.T) {
	source := "let x 5;"

	diag := DiagnosticError{
		Line:    1,
		Column:  7,
		Length:  1,
		Message: "expected next token to be =, got INT instead",
	}

	result := FormatDiagnostic(source, diag)

	// 基本構造の検証
	if !strings.Contains(result, "error:") {
		t.Error("should contain 'error:' label")
	}
	if !strings.Contains(result, "1:7") {
		t.Error("should contain position '1:7'")
	}
	if !strings.Contains(result, "let x 5;") {
		t.Error("should contain source line")
	}
	if !strings.Contains(result, "^") {
		t.Error("should contain caret marker")
	}
}

func TestFormatDiagnosticMultiLine(t *testing.T) {
	source := "let x = 10;\nlet y 20;\nlet z = 30;"

	diag := DiagnosticError{
		Line:    2,
		Column:  7,
		Length:  2,
		Message: "expected =, got INT",
	}

	result := FormatDiagnostic(source, diag)

	if !strings.Contains(result, "2:7") {
		t.Error("should reference line 2, column 7")
	}
	if !strings.Contains(result, "let y 20;") {
		t.Error("should show the correct source line")
	}
	if !strings.Contains(result, "^^") {
		t.Error("should have double caret for length 2")
	}
}

func TestFormatDiagnosticWithHint(t *testing.T) {
	source := "let x == 5;"

	diag := DiagnosticError{
		Line:    1,
		Column:  7,
		Length:  2,
		Message: "expected =, got ==",
		Hint:    "use '=' for assignment, not '=='",
	}

	result := FormatDiagnostic(source, diag)

	if !strings.Contains(result, "hint:") {
		t.Error("should contain hint")
	}
	if !strings.Contains(result, "use '=' for assignment") {
		t.Error("should contain hint message")
	}
}

func TestFormatDiagnosticColor(t *testing.T) {
	source := "let x 5;"

	diag := DiagnosticError{
		Line:    1,
		Column:  7,
		Length:  1,
		Message: "unexpected token",
	}

	result := FormatDiagnosticColor(source, diag)

	// Should contain ANSI escape codes
	if !strings.Contains(result, "\033[") {
		t.Error("color version should contain ANSI escape codes")
	}
}

func TestFormatSimpleErrors(t *testing.T) {
	source := "let x 5;"
	errors := []string{
		"expected next token to be =, got INT instead",
	}

	result := FormatSimpleErrors(source, errors)

	if !strings.Contains(result, "error:") {
		t.Error("should contain error label")
	}
}

func TestFormatDiagnosticEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
		diag   DiagnosticError
		want   []string // strings that should be present
	}{
		{
			name:   "column at start",
			source: "x = 5;",
			diag: DiagnosticError{
				Line:    1,
				Column:  1,
				Length:  1,
				Message: "unexpected identifier",
			},
			want: []string{"error:", "1:1", "^"},
		},
		{
			name:   "long underline",
			source: "let variable = 10;",
			diag: DiagnosticError{
				Line:    1,
				Column:  5,
				Length:  8,
				Message: "invalid variable name",
			},
			want: []string{"error:", "1:5", "^^^^^^^^"},
		},
		{
			name:   "line out of range",
			source: "let x = 5;",
			diag: DiagnosticError{
				Line:    10,
				Column:  1,
				Length:  1,
				Message: "line out of range",
			},
			want: []string{"error:", "10:1"},
		},
		{
			name:   "zero column defaults to 1",
			source: "let x = 5;",
			diag: DiagnosticError{
				Line:    1,
				Column:  0,
				Length:  1,
				Message: "test",
			},
			want: []string{"error:", "1:0", "^"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDiagnostic(tt.source, tt.diag)
			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("FormatDiagnostic() missing %q in output:\n%s", want, result)
				}
			}
		})
	}
}
