package errs

import (
	"github.com/CanPacis/go-i18n/parser/token"
)

type Positioner interface {
	Position() (start token.Position, end token.Position)
}

type TopLevelError interface {
	error
	Positioner
	Unwrap() []error
	File() string
}
