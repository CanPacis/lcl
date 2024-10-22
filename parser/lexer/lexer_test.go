package lexer_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/parser/lexer"
	"github.com/CanPacis/go-i18n/parser/token"
	"github.com/stretchr/testify/assert"
)

type Expectation struct {
	Kind    token.Kind
	Start   token.Position
	End     token.Position
	Literal string
}

type Case struct {
	Input           string
	Expected        []Expectation
	skipsWhitespace bool
}

type CaseList []Case

func (cl CaseList) Run(t *testing.T) {
	assert := assert.New(t)

	for i, test := range cl {
		buf := bytes.NewBuffer([]byte(test.Input))
		l := lexer.New(buf)

		var tk token.Token
		var j = 0
		for tk.Kind != token.EOF {
			tk = l.Next()
			if tk.Kind == token.WHITESPACE && test.skipsWhitespace {
				continue
			}
			if tk.Kind == token.EOF {
				continue
			}

			if len(test.Expected) < j+1 {
				assert.Equal(len(test.Expected), j+1, "Got more tokens than expected")
			}

			expected := test.Expected[j]
			assert.Equal(expected.Kind.String(), tk.Kind.String(), msg("Kind", i, j))
			assert.Equal(expected.Start.Line, tk.Start.Line, msg("Start line", i, j))
			assert.Equal(expected.Start.Column, tk.Start.Column, msg("Start column", i, j))
			assert.Equal(expected.End.Line, tk.End.Line, msg("End line", i, j))
			assert.Equal(expected.End.Column, tk.End.Column, msg("End column", i, j))
			assert.Equal(expected.Literal, tk.Literal, msg("Literal", i, j))

			j++
		}
	}
}

func Exp(kind token.Kind, literal string, sl, sc, el, ec int) Expectation {
	return Expectation{
		Kind:    kind,
		Literal: literal,
		Start:   token.NewPosition(sl, sc),
		End:     token.NewPosition(el, ec),
	}
}

func msg(field string, i, j int) string {
	return fmt.Sprintf("Case %d expectation %d: %s failed", i, j, field)
}

func TestSpecialChars(t *testing.T) {
	tests := CaseList{
		{
			skipsWhitespace: true,
			Input:           "( ) { } [ ] . , ::: ? == != > >= < <= *",
			Expected: []Expectation{
				Exp(token.LEFT_PARENS, "(", 1, 1, 1, 2),
				Exp(token.RIGHT_PARENS, ")", 1, 3, 1, 4),
				Exp(token.LEFT_CURLY_BRACE, "{", 1, 5, 1, 6),
				Exp(token.RIGHT_CURLY_BRACE, "}", 1, 7, 1, 8),
				Exp(token.LEFT_SQUARE_BRACKET, "[", 1, 9, 1, 10),
				Exp(token.RIGHT_SQUARE_BRACKET, "]", 1, 11, 1, 12),
				Exp(token.DOT, ".", 1, 13, 1, 14),
				Exp(token.COMMA, ",", 1, 15, 1, 16),
				Exp(token.DOUBLE_COLON, "::", 1, 17, 1, 19),
				Exp(token.COLON, ":", 1, 19, 1, 20),
				Exp(token.QUESTION_MARK, "?", 1, 21, 1, 22),
				Exp(token.EQUALS, "==", 1, 23, 1, 25),
				Exp(token.NOT_EQUALS, "!=", 1, 26, 1, 28),
				Exp(token.GT, ">", 1, 29, 1, 30),
				Exp(token.GTE, ">=", 1, 31, 1, 33),
				Exp(token.LT, "<", 1, 34, 1, 35),
				Exp(token.LTE, "<=", 1, 36, 1, 38),
				Exp(token.STAR, "*", 1, 39, 1, 40),
			},
		},
	}

	tests.Run(t)
}

func TestAlphanumeric(t *testing.T) {
	tests := CaseList{
		{
			skipsWhitespace: true,
			Input:           "for in import proc type section key template identifier",
			Expected: []Expectation{
				Exp(token.FOR, "for", 1, 1, 1, 4),
				Exp(token.IN, "in", 1, 5, 1, 7),
				Exp(token.IMPORT, "import", 1, 8, 1, 14),
				Exp(token.PROC, "proc", 1, 15, 1, 19),
				Exp(token.TYPE, "type", 1, 20, 1, 24),
				Exp(token.SECTION, "section", 1, 25, 1, 32),
				Exp(token.KEY, "key", 1, 33, 1, 36),
				Exp(token.TEMPLATE, "template", 1, 37, 1, 45),
				Exp(token.IDENT, "identifier", 1, 46, 1, 56),
			},
		},
	}

	tests.Run(t)
}

func TestString(t *testing.T) {
	tests := CaseList{
		{
			Input: `"this is some string"`,
			Expected: []Expectation{
				Exp(token.STRING, "this is some string", 1, 1, 1, 22),
			},
		},
		{
			Input: `"esca\"ped string"`,
			Expected: []Expectation{
				Exp(token.STRING, `esca"ped string`, 1, 1, 1, 19),
			},
		},
		{
			Input: `"unterminated string`,
			Expected: []Expectation{
				Exp(token.UNTERM_STR, `"unterminated string`, 1, 1, 1, 21),
			},
		},
		{
			Input: "`template string`",
			Expected: []Expectation{
				Exp(token.TEMPLATE_LIT, "template string", 1, 1, 1, 18),
			},
		},
		{
			Input: "`unterminated template",
			Expected: []Expectation{
				Exp(token.UNTERM_TEMP, "`unterminated template", 1, 1, 1, 23),
			},
		},
		{
			Input: "`template with { expression \"\" `` }`",
			Expected: []Expectation{
				Exp(token.TEMPLATE_LIT, "template with { expression \"\" `` }", 1, 1, 1, 37),
			},
		},
		{
			// the expression should fail, the last pair of `'s form a valid template
			// so the first expression is left unclosed
			Input: "`unterminated template { inside expr ` }`",
			Expected: []Expectation{
				Exp(token.UNTERM_TEMP_EXPR, "{", 1, 24, 1, 25),
			},
		},
	}

	tests.Run(t)
}

func TestComment(t *testing.T) {
	tests := CaseList{
		{
			skipsWhitespace: true,
			Input:           "# test comment",
			Expected: []Expectation{
				Exp(token.COMMENT, " test comment", 1, 1, 1, 15),
			},
		},
		{
			skipsWhitespace: true,
			Input:           "# test comment\n\n#another comment",
			Expected: []Expectation{
				Exp(token.COMMENT, " test comment", 1, 1, 1, 15),
				Exp(token.COMMENT, "another comment", 3, 1, 3, 17),
			},
		},
	}

	tests.Run(t)
}

func TestNumber(t *testing.T) {
	tests := CaseList{
		{
			skipsWhitespace: true,
			Input:           "- 0 -3 04 0. 0.3 245 36.6 -2 -0.1 -0",
			Expected: []Expectation{
				Exp(token.ILLEGAL, "-", 1, 1, 1, 2),
				Exp(token.NUMBER, "0", 1, 3, 1, 4),
				Exp(token.NUMBER, "-3", 1, 5, 1, 7),
				Exp(token.ILLEGAL, "04", 1, 8, 1, 10),
				Exp(token.ILLEGAL, "0.", 1, 11, 1, 13),
				Exp(token.NUMBER, "0.3", 1, 14, 1, 17),
				Exp(token.NUMBER, "245", 1, 18, 1, 21),
				Exp(token.NUMBER, "36.6", 1, 22, 1, 26),
				Exp(token.NUMBER, "-2", 1, 27, 1, 29),
				Exp(token.NUMBER, "-0.1", 1, 30, 1, 34),
				Exp(token.NUMBER, "-0", 1, 35, 1, 37),
			},
		},
	}

	tests.Run(t)
}
