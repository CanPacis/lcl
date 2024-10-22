package errs

import (
	"fmt"

	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
)

type SemanticError struct {
	Reasons []error
	file    string
}

func (e *SemanticError) Position() (start token.Position, end token.Position) {
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

func (e *SemanticError) File() string {
	return e.file
}

func NewSemanticError(reasons []error, file string) *SemanticError {
	return &SemanticError{
		Reasons: reasons,
		file:    file,
	}
}

type Resolvable string

const (
	IMPORT Resolvable = "import"
	TARGET Resolvable = "target"
	TYPE   Resolvable = "type"
	PROC   Resolvable = "proc"
	CONST  Resolvable = "const"
)

const (
	Unresolved = "unresolved"
	Duplicate  = "duplicate definition"
	Type       = "type error"

	NotComparable = "expressions are not comparable"
	NotInferrable = "expression's type cannot be inferred"
	PredIsNonBool = "predicate expression must be a bool"
	PredIsInvalid = "both sides of the predicate must be the same type"
	NotCallable   = "expression is not callable"
	NotAssignable = "expression is not assignable"
	NotIndexable  = "expression is not indexable"
)

type ResolveError struct {
	Value string
	Kind  Resolvable
	Node  ast.Node
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("%s %s: %s", Unresolved, e.Kind, e.Value)
}

func (e *ResolveError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type TypeError struct {
	Message string
	Node    ast.Node
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("%s: %s", Type, e.Message)
}

func (e *TypeError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

func NewTypeError(node ast.Node, message string, a ...any) *TypeError {
	return &TypeError{
		Message: fmt.Sprintf(message, a...),
		Node:    node,
	}
}

type DuplicateDefError struct {
	Name     string
	Original ast.Node
	Node     ast.Node
}

func (e *DuplicateDefError) Error() string {
	return fmt.Sprintf(
		"%s: '%s' is already defined here %s - %s",
		Duplicate,
		e.Name,
		e.Original.Start(),
		e.Original.End(),
	)
}

func (e *DuplicateDefError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}
