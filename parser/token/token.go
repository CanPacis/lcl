package token

import (
	"encoding/json"
	"fmt"
)

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Range struct {
	Start Position
	End   Position
}

func (r Range) String() string {
	return fmt.Sprintf("%s - %s", r.Start, r.End)
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

func (k Kind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
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
	PLUS
	MINUS
	FORWARD_SLASH
	PERCENT
	CARET

	// Delimeter
	LOGICAL

	AND
	OR
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
	TEMPLATE
	NUMBER

	// Delimeter
	KEYWORD

	DECLARE
	AS
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
	PLUS:                 "+",
	MINUS:                "-",
	FORWARD_SLASH:        "/",
	PERCENT:              "%",

	LOGICAL: "logical",

	AND:        "&&",
	OR:         "||",
	EQUALS:     "==",
	NOT_EQUALS: "!=",
	GT:         ">",
	GTE:        ">=",
	LT:         "<",
	LTE:        "<=",

	LITERAL: "literal",

	IDENT:    "identifier",
	STRING:   "string literal",
	TEMPLATE: "template literal",
	NUMBER:   "number literal",

	KEYWORD: "keyword",

	DECLARE: "declare",
	AS:      "as",
	IMPORT:  "import",
	FN:      "fn",
	TYPE:    "type",
	SECTION: "section",
}

type Token struct {
	Kind    Kind     `json:"kind"`
	Literal string   `json:"literal"`
	Raw     string   `json:"raw"`
	Start   Position `json:"start"`
	End     Position `json:"end"`
}

func (t Token) String() string {
	if t.Kind > PUNCTUATION && t.Kind < LOGICAL {
		return fmt.Sprintf("%s %s", PUNCTUATION.String(), t.Kind.String())
	}
	if t.Kind > LOGICAL && t.Kind < LITERAL {
		return fmt.Sprintf("%s %s", LOGICAL.String(), t.Kind.String())
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

func (t Token) Range() Range {
	return Range{
		Start: t.Start,
		End:   t.End,
	}
}
