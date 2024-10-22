package token

import "fmt"

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func NewPosition(line, col int) Position {
	return Position{
		Line:   line,
		Column: col,
	}
}

type Kind int

func (k Kind) String() string {
	if tokenMap[k] == "" {
		return tokenMap[UNKNOWN]
	}
	return tokenMap[k]
}

const (
	// Control tokens

	ILLEGAL Kind = iota
	UNKNOWN
	UNTERM_STR
	UNTERM_TEMP
	UNTERM_TEMP_EXPR
	EOF

	WHITESPACE
	COMMENT

	// Delimeter
	PUNCTUATION

	LEFT_PARENS
	RIGHT_PARENS
	LEFT_CURLY_BRACE
	RIGHT_CURLY_BRACE
	LEFT_SQUARE_BRACKET
	RIGHT_SQUARE_BRACKET
	DOT
	COMMA
	COLON
	DOUBLE_COLON
	QUESTION_MARK
	STAR

	// Delimeter
	ARITHMETIC

	EQUALS
	NOT_EQUALS
	GT
	GTE
	LT
	LTE

	// Delimeter
	LITERAL

	IDENT
	STRING
	TEMPLATE_LIT
	NUMBER

	// Delimeter
	KEYWORD

	FOR
	IN
	IMPORT
	FN
	TYPE
	SECTION
)

var tokenMap = map[Kind]string{
	ILLEGAL:          "illegal",
	UNKNOWN:          "unknown",
	UNTERM_STR:       "string",
	UNTERM_TEMP:      "template",
	UNTERM_TEMP_EXPR: "template expression",
	EOF:              "eof",

	WHITESPACE: "whitespace",
	COMMENT:    "comment",

	PUNCTUATION: "punctuation",

	LEFT_PARENS:          "(",
	RIGHT_PARENS:         ")",
	LEFT_CURLY_BRACE:     "{",
	RIGHT_CURLY_BRACE:    "}",
	LEFT_SQUARE_BRACKET:  "[",
	RIGHT_SQUARE_BRACKET: "]",
	DOT:                  ".",
	COMMA:                ",",
	COLON:                ":",
	DOUBLE_COLON:         "::",
	QUESTION_MARK:        "?",
	STAR:                 "*",

	ARITHMETIC: "arithmetic",

	EQUALS:     "==",
	NOT_EQUALS: "!=",
	GT:         ">",
	GTE:        ">=",
	LT:         "<",
	LTE:        "<=",

	LITERAL: "literal",

	IDENT:        "identifier",
	STRING:       "string literal",
	TEMPLATE_LIT: "template literal",
	NUMBER:       "number literal",

	KEYWORD: "keyword",

	FOR:     "for",
	IN:      "in",
	IMPORT:  "import",
	FN:      "fn",
	TYPE:    "type",
	SECTION: "section",
}

type Token struct {
	Kind    Kind
	Literal string
	Raw     string
	Start   Position
	End     Position
}

func (t Token) String() string {
	if t.Kind > PUNCTUATION && t.Kind < ARITHMETIC {
		return fmt.Sprintf("%s %s", PUNCTUATION.String(), t.Kind.String())
	}
	if t.Kind > ARITHMETIC && t.Kind < LITERAL {
		return fmt.Sprintf("%s %s", ARITHMETIC.String(), t.Kind.String())
	}
	if t.Kind > KEYWORD {
		return fmt.Sprintf("%s <%s>", KEYWORD.String(), t.Literal)
	}

	switch t.Kind {
	case WHITESPACE:
		return "whitespace"
	case EOF:
		return "eof"
	default:
		return fmt.Sprintf("%s (%s)", t.Kind.String(), t.Literal)
	}
}
