package parser

import (
	"fmt"
	"strings"
)

// STATUS: 未統合の実験コード（frozen）。
//
// Rust スタイルの診断表示を提供するが、パーサー本体からは呼ばれていない。
// 本統合には token.Token への行・列情報の追加と lexer/parser の改修が必要
// （ROADMAP Phase 1.2 相当）。着手判断が付くまで、テスト付きの独立した部品
// として凍結保存する。削除する場合は diagnostic_test.go ごと除去してよい。

// DiagnosticError はソースコード位置付きのエラー表示情報を持つ
type DiagnosticError struct {
	Line    int
	Column  int
	Length  int // エラー箇所の長さ（rune単位）
	Message string
	Hint    string // オプション: ヒントメッセージ
}

// ANSI color constants
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// FormatDiagnostic はRustスタイルのエラーメッセージを生成する（カラーなし）
func FormatDiagnostic(source string, diag DiagnosticError) string {
	return formatDiagnosticImpl(source, diag, false)
}

// FormatDiagnosticColor はRustスタイルのカラー付きエラーメッセージを生成する
func FormatDiagnosticColor(source string, diag DiagnosticError) string {
	return formatDiagnosticImpl(source, diag, true)
}

func formatDiagnosticImpl(source string, diag DiagnosticError, color bool) string {
	var b strings.Builder
	lines := strings.Split(source, "\n")

	// エラーラベル
	if color {
		b.WriteString(colorRed + colorBold + "error" + colorReset)
		b.WriteString(colorBold + ": " + diag.Message + colorReset)
	} else {
		b.WriteString("error: " + diag.Message)
	}
	b.WriteString("\n")

	// 位置情報
	locStr := fmt.Sprintf("  --> %d:%d", diag.Line, diag.Column)
	if color {
		b.WriteString(colorCyan + locStr + colorReset)
	} else {
		b.WriteString(locStr)
	}
	b.WriteString("\n")

	// ソースコード取得
	if diag.Line < 1 || diag.Line > len(lines) {
		return b.String()
	}

	lineContent := lines[diag.Line-1]
	lineNumStr := fmt.Sprintf("%d", diag.Line)
	padding := strings.Repeat(" ", len(lineNumStr))

	// 空行
	if color {
		b.WriteString(colorCyan + padding + " |" + colorReset + "\n")
	} else {
		b.WriteString(padding + " |\n")
	}

	// ソースコード行
	if color {
		b.WriteString(colorCyan + lineNumStr + " | " + colorReset + lineContent + "\n")
	} else {
		b.WriteString(lineNumStr + " | " + lineContent + "\n")
	}

	// エラー指示行（^^^^^）
	col := diag.Column
	if col < 1 {
		col = 1
	}
	length := diag.Length
	if length < 1 {
		length = 1
	}
	underline := strings.Repeat(" ", col-1) + strings.Repeat("^", length)
	if color {
		b.WriteString(colorCyan + padding + " | " + colorReset + colorRed + underline + colorReset + "\n")
	} else {
		b.WriteString(padding + " | " + underline + "\n")
	}

	// ヒント
	if diag.Hint != "" {
		if color {
			b.WriteString(colorCyan + padding + " = " + colorReset + colorBlue + "hint: " + diag.Hint + colorReset + "\n")
		} else {
			b.WriteString(padding + " = hint: " + diag.Hint + "\n")
		}
	}

	return b.String()
}

// FormatSimpleErrors は簡易的な位置付きエラーメッセージを生成（[]string互換）
// source: ソースコード文字列
// errors: parser.Errors()が返す[]string（"expected next token..."形式）
// この関数はParseError型なしでも動作する（後方互換）
func FormatSimpleErrors(source string, errors []string) string {
	var b strings.Builder
	for _, msg := range errors {
		b.WriteString("error: " + msg + "\n")
	}
	return b.String()
}
