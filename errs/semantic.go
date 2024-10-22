package errs

import (
	"fmt"
	"strings"

	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
)

type SemanticError struct {
	Reasons []error
	File    string
}

func (e *SemanticError) Position() (start token.Position, end token.Position) {
	if len(e.Reasons) == 0 {
		return
	}

	reason := e.Reasons[0]
	p, ok := reason.(position)
	if !ok {
		return
	}

	return p.Position()
}

func (e *SemanticError) Error() string {
	if len(e.Reasons) == 0 {
		return ""
	}
	reason := e.Reasons[0]

	return fmt.Sprintf("semantic error: %s", reason.Error())
}

func (e *SemanticError) Unwrap() []error {
	return e.Reasons
}

type Resolvable string

const (
	IMPORT Resolvable = "import"
	TARGET Resolvable = "target"
	TYPE   Resolvable = "type"
	PROC   Resolvable = "proc"
	CONST  Resolvable = "const"
)

type ResolveError struct {
	Value string
	Kind  Resolvable
	Node  ast.Node
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("unresolved %s: %s", e.Kind, e.Value)
}

func (e *ResolveError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type DuplicateDefError struct {
	Name     string
	Original ast.Node
	Node     ast.Node
}

func (e *DuplicateDefError) Error() string {
	return fmt.Sprintf(
		"duplicate definition: '%s' is already defined here %s - %s",
		e.Name,
		e.Original.Start(),
		e.Original.End(),
	)
}

func (e *DuplicateDefError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type FieldDeclError struct {
	Excess  []string
	Missing []string
	Entry   ast.Entry
}

func (e *FieldDeclError) Error() string {
	return fmt.Sprintf(
		"field declaration: missing %s and excess %s",
		strings.Join(e.Missing, " "),
		strings.Join(e.Excess, " "),
	)
}

func (e *FieldDeclError) Position() (start token.Position, end token.Position) {
	return e.Entry.Start(), e.Entry.End()
}
