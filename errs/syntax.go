package errs

import (
	"fmt"
	"strings"

	"github.com/CanPacis/go-i18n/parser/token"
)

type SyntaxError struct {
	Reasons []error
	file    string
}

func (e *SyntaxError) Position() (start token.Position, end token.Position) {
	if len(e.Reasons) == 0 {
		return
	}

	reason := e.Reasons[0]
	p, ok := reason.(Positioner)
	if !ok {
		return
	}

	return p.Position()
}

func (e *SyntaxError) Error() string {
	if len(e.Reasons) == 0 {
		return ""
	}
	reason := e.Reasons[0]

	return fmt.Sprintf("syntax error: %s", reason.Error())
}

func (e *SyntaxError) Unwrap() []error {
	return e.Reasons
}

func (e *SyntaxError) File() string {
	return e.file
}

func NewSyntaxError(reasons []error, file string) *SyntaxError {
	return &SyntaxError{
		Reasons: reasons,
		file:    file,
	}
}

type UnexpectedTokenError struct {
	Details  string
	Found    token.Token
	Expected []token.Kind
}

func (e *UnexpectedTokenError) Position() (start token.Position, end token.Position) {
	return e.Found.Start, e.Found.End
}

func (e *UnexpectedTokenError) Error() string {
	details := ""

	if len(e.Details) > 0 {
		details = ", " + e.Details
	}

	expected := []string{}
	for _, e := range e.Expected {
		expected = append(expected, fmt.Sprintf("'%s'", e.String()))
	}

	if len(expected) > 0 {
		if len(expected) == 1 {
			return fmt.Sprintf("unexpected token: '%s', was expecting a %s%s", e.Found.Kind.String(), expected[0], details)
		}
		return fmt.Sprintf(
			"unexpected token: '%s', was expecting any of %s%s",
			e.Found.Kind.String(),
			strings.Join(expected, ", "),
			details,
		)
	}

	return fmt.Sprintf("unexpected token: '%s'%s", e.Found.Kind.String(), details)
}

type UntermConstructError struct {
	Token token.Token
}

func (e *UntermConstructError) Error() string {
	return fmt.Sprintf("unterminated %s: token does not have an ending", e.Token.Kind.String())
}

func (e *UntermConstructError) Position() (start token.Position, end token.Position) {
	return e.Token.Start, e.Token.End
}

type NumberError struct {
	Reason error
	Token  token.Token
}

func (e *NumberError) Position() (start token.Position, end token.Position) {
	return e.Token.Start, e.Token.End
}

func (e *NumberError) Error() string {
	if e.Reason == nil {
		return ""
	}

	return fmt.Sprintf("number error: %s", e.Reason.Error())
}

func (e *NumberError) Unwrap() error {
	return e.Reason
}
