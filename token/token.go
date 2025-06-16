package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	// 識別子 + リテラル
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"   // 1343456
	FLOAT = "FLOAT" // 13.45
	// 演算子
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT  = "<"
	GT  = ">"
	LTE = "<="
	GTE = ">="
	AND = "&&"
	OR  = "||"
	// デリミタ
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	// キーワード
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	EQ       = "=="
	NOT_EQ   = "!="

	// 代入演算子
	PLUS_ASSIGN     = "+="
	MINUS_ASSIGN    = "-="
	ASTERISK_ASSIGN = "*="
	SLASH_ASSIGN    = "/="

	STRING = "STRING"

	LBRACKET = "["
	RBRACKET = "]"

	COLON = ":"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
