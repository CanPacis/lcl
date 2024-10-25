package lexer

import (
	"bufio"
	"bytes"
	"io"
	"slices"
	"strings"
	"unicode"

	"github.com/CanPacis/lcl/parser/token"
)

const eof rune = -1

var (
	keywords = map[string]token.Kind{
		"declare": token.DECLARE,
		"as":      token.AS,
		"import":  token.IMPORT,
		"fn":      token.FN,
		"type":    token.TYPE,
		"section": token.SECTION,
	}

	special = []rune{
		'{', '}', '(', ')', '.', ';', ',', '[', ']', '-', '+', '/',
		'>', '<', '=', ':', '?', '!', '|', '`', '*', '&', '_', '%', '^',
	}

	EOF = token.Token{
		Kind:    token.EOF,
		Literal: "",
		Raw:     "",
		Start:   token.Position{},
		End:     token.Position{},
	}
)

type Lexer struct {
	input *bufio.Reader

	current    rune
	past       []rune
	newLine    bool
	start, end token.Position

	// last extracted tokens from a template
	template []token.Token
}

func (l *Lexer) init() {
	l.current = l.next()
}

func (l *Lexer) next() rune {
	r, _, err := l.input.ReadRune()
	if err != nil {
		return eof
	}

	return r
}

func (l *Lexer) last() rune {
	if len(l.past) == 0 {
		return eof
	}

	return l.past[len(l.past)-1]
}

func (l *Lexer) advance() rune {
	rn := l.next()

	if l.newLine {
		l.end.Line++
		l.end.Column = 0
		l.newLine = false
	}

	if rn == '\n' || rn == '\r' {
		l.newLine = true
	}

	l.end.Column++

	l.past = append(l.past, l.current)
	l.current = rn
	return rn
}

func (l *Lexer) word() string {
	return string(l.past)
}

func (l *Lexer) token(kind token.Kind) token.Token {
	raw := l.word()
	literal := raw

	if kind == token.STRING || kind == token.TEMPLATE {
		literal = raw[1 : len(raw)-1]
		literal = strings.ReplaceAll(literal, "\\\"", "\"")
	}
	if kind == token.COMMENT {
		literal = raw[1:]
	}

	tk := token.Token{
		Kind:    kind,
		Literal: literal,
		Raw:     raw,
		Start:   l.start,
		End:     l.end,
	}
	l.start = l.end
	l.past = []rune{}
	return tk
}

func (l *Lexer) Next() token.Token {
	switch l.current {
	case '"':
		return l.lexString()
	case '`':
		return l.lexTemplateLit()
	case '#':
		return l.lexComment()
	case eof:
		return EOF
	default:
		switch {
		case unicode.IsLetter(l.current):
			return l.lexAlphanumeric()
		case unicode.IsDigit(l.current):
			return l.lexNumber()
		case unicode.IsSpace(l.current):
			return l.lexSpace()
		case slices.Contains(special, l.current):
			return l.lexSpecial()
		default:
			tk := l.token(token.ILLEGAL)
			l.advance()
			return tk
		}
	}
}

func (l *Lexer) lexSpecial() token.Token {
	var tk token.Token

	switch l.current {
	case '(':
		l.advance()
		tk = l.token(token.LEFT_PARENS)
	case ')':
		l.advance()
		tk = l.token(token.RIGHT_PARENS)
	case '{':
		l.advance()
		tk = l.token(token.LEFT_CURLY_BRACE)
	case '}':
		l.advance()
		tk = l.token(token.RIGHT_CURLY_BRACE)
	case '[':
		l.advance()
		tk = l.token(token.LEFT_SQUARE_BRACKET)
	case ']':
		l.advance()
		tk = l.token(token.RIGHT_SQUARE_BRACKET)
	case '.':
		l.advance()
		tk = l.token(token.DOT)
	case ',':
		l.advance()
		tk = l.token(token.COMMA)
	case '^':
		l.advance()
		tk = l.token(token.CARET)
	case '*':
		l.advance()
		return l.token(token.STAR)
	case '+':
		l.advance()
		return l.token(token.PLUS)
	case '-':
		l.advance()

		if unicode.IsDigit(l.current) {
			return l.lexNumber()
		} else {
			tk = l.token(token.MINUS)
		}
	case '/':
		l.advance()
		return l.token(token.FORWARD_SLASH)
	case '%':
		l.advance()
		return l.token(token.PERCENT)
	case ':':
		l.advance()

		if l.current == ':' {
			l.advance()
			tk = l.token(token.DOUBLE_COLON)
		} else {
			tk = l.token(token.COLON)
		}
	case '?':
		l.advance()
		tk = l.token(token.QUESTION_MARK)
	case '=':
		l.advance()

		if l.current == '=' {
			l.advance()
			tk = l.token(token.EQUALS)
		}
	case '!':
		l.advance()

		if l.current == '=' {
			l.advance()
			tk = l.token(token.NOT_EQUALS)
		}
	case '&':
		l.advance()

		if l.current == '&' {
			l.advance()
			tk = l.token(token.AND)
		}
	case '|':
		l.advance()

		if l.current == '|' {
			l.advance()
			tk = l.token(token.OR)
		}
	case '>':
		l.advance()

		if l.current == '=' {
			l.advance()
			tk = l.token(token.GTE)
		} else {
			tk = l.token(token.GT)
		}
	case '<':
		l.advance()

		if l.current == '=' {
			l.advance()
			tk = l.token(token.LTE)
		} else {
			tk = l.token(token.LT)
		}
	default:
		l.advance()
	}

	return tk
}

func (l *Lexer) lexAlphanumeric() token.Token {
	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		l.advance()
	}

	keyword := keywords[l.word()]
	if keyword != token.ILLEGAL {
		return l.token(keyword)
	} else {
		return l.token(token.IDENT)
	}
}

func (l *Lexer) lexNumber() token.Token {
	if l.last() == '-' {
		if !unicode.IsDigit(l.current) {
			return l.token(token.ILLEGAL)
		}
	}

	if l.current == '0' {
		l.advance()

		if unicode.IsDigit(l.current) {
			for unicode.IsDigit(l.current) {
				l.advance()
			}
			return l.token(token.ILLEGAL)
		}
	}

	for unicode.IsDigit(l.current) {
		l.advance()
	}

	if l.current == '.' {
		l.advance()

		if !unicode.IsDigit(l.current) {
			return l.token(token.ILLEGAL)
		} else {
			for unicode.IsDigit(l.current) {
				l.advance()
			}
		}
	}

	return l.token(token.NUMBER)
}

func (l *Lexer) lexString() token.Token {
	l.advance()

	for l.current != eof && l.current != '\n' && l.current != '\r' {
		if l.current == '"' && l.last() != '\\' {
			break
		}

		l.advance()
	}

	if l.current != '"' {
		return l.token(token.UNTERM_STR)
	}
	l.advance()

	return l.token(token.STRING)
}

func joinTokens(list []token.Token, kind token.Kind) token.Token {
	raw := ""
	for _, token := range list {
		raw += token.Raw
	}

	return token.Token{
		Kind:    kind,
		Literal: raw[1 : len(raw)-1],
		Raw:     raw,
		Start:   list[0].Start,
		End:     list[len(list)-1].End,
	}
}

func (l *Lexer) lexTemplateLit() token.Token {
	l.advance()

	// will accumulate the tokens inside and join them
	tokens := []token.Token{}

	for l.current != eof && l.current != '`' {
		if l.current == '{' && l.last() != '\\' {
			tokens = append(tokens, l.token(token.UNKNOWN))
			l.advance()
			start := l.token(token.UNTERM_TEMP_EXPR)
			tokens = append(tokens, start)

			var tk token.Token
			for tk.Kind != token.EOF && tk.Kind != token.RIGHT_CURLY_BRACE {
				tk = l.Next()
				tokens = append(tokens, tk)
			}

			if tokens[len(tokens)-1].Kind != token.RIGHT_CURLY_BRACE {
				return start
			}

		} else {
			l.advance()
		}
	}

	if l.current != '`' {
		return l.token(token.UNTERM_TEMP)
	}
	l.advance()
	tokens = append(tokens, l.token(token.UNKNOWN))
	l.template = tokens
	return joinTokens(tokens, token.TEMPLATE)
}

func (l *Lexer) lexSpace() token.Token {
	for unicode.IsSpace(l.current) {
		l.advance()
	}
	return l.token(token.WHITESPACE)
}

func (l *Lexer) lexComment() token.Token {
	for l.current != eof && l.current != '\n' && l.current != '\r' {
		l.advance()
	}
	return l.token(token.COMMENT)
}

func LexTemplate(start token.Token) []token.Token {
	lexer := New(bytes.NewBuffer([]byte(start.Raw)))
	lexer.start = start.Start
	lexer.end = start.Start

	for lexer.current != eof {
		lexer.Next()
	}
	return lexer.template
}

func New(input io.Reader) *Lexer {
	lexer := &Lexer{
		input:   bufio.NewReader(input),
		current: eof,
		start:   token.NewPosition(1, 1),
		end:     token.NewPosition(1, 1),
	}
	lexer.init()
	return lexer
}
