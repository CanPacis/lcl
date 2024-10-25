package errs

import (
	"errors"
	"fmt"

	"github.com/CanPacis/lcl/parser/token"
)

var (
	ErrUnexpectedToken       = errors.New("unexpected token")
	ErrUnterminatedConstruct = errors.New("unterminated")
	ErrMalformedNumber       = errors.New("malformed number")
)

func unexpectedToken(t token.Token, e []token.Kind, d string) string {
	if len(d) > 0 {
		d = ", " + d
	}
	base := fmt.Sprintf("%s '%s'", ErrUnexpectedToken.Error(), t.Kind.String())

	if len(e) == 0 {
		return base
	}
	switch len(e) {
	case 0:
		return base
	case 1:
		return fmt.Sprintf("%s, was expecting a %s%s", base, e[0].String(), d)
	default:
		return fmt.Sprintf("%s, was expecting any of %s%s", base, join(e, ", "), d)
	}
}

type SyntaxError struct {
	Err      error
	Token    token.Token
	Expected []token.Kind
	Details  string
}

func (e *SyntaxError) Error() string {
	switch {
	case errors.Is(e.Err, ErrUnexpectedToken):
		return fmt.Sprintf("%s: %s", e.Name(), unexpectedToken(e.Token, e.Expected, e.Details))
	case errors.Is(e.Err, ErrUnterminatedConstruct):
		return fmt.Sprintf("%s: %s %s, token does not have an ending", e.Name(), e.Err.Error(), e.Token.Kind.String())
	case errors.Is(e.Err, ErrMalformedNumber):
		return fmt.Sprintf("%s: %s", e.Name(), e.Err.Error())
	default:
		return fmt.Sprintf("%s: %s", e.Name(), e.Err.Error())
	}
}

func (e *SyntaxError) Unwrap() error {
	return e.Err
}

func (e *SyntaxError) Name() string {
	return "syntax error"
}

func (e *SyntaxError) Range() token.Range {
	return e.Token.Range()
}
