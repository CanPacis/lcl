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

// type UnsupportedError struct {
// 	Node ast.Node
// }

// func (e *UnsupportedError) Error() string {
// 	return "unsupported"
// }

// func (e *UnsupportedError) Position() (start token.Position, end token.Position) {
// 	return e.Node.Start(), e.Node.End()
// }
