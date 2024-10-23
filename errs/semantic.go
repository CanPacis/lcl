package errs

import (
	"fmt"

	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
	"golang.org/x/text/language"
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

	return "semantic error: " + reason.Error()
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

const (
	Unresolved     = "unresolved"
	Duplicate      = "duplicate definition"
	Type           = "type error"
	TargetMismatch = "target mismatch"

	NotComparable = "expressions are not comparable"
	PredIsNonBool = "predicate expression must be a bool"
	PredIsInvalid = "both sides of the predicate must be the same type"
	ArgumentCount = "invalid number of arguments"
	NotCallable   = "expression is not callable"
	NotAssignable = "expression is not assignable"
	NotIndexable  = "expression is not indexable"
)

type Resolvable string

const (
	IMPORT Resolvable = "import"
	TARGET Resolvable = "target"
	TAG    Resolvable = "tag"
	TYPE   Resolvable = "type"
	FN     Resolvable = "fn"
	CONST  Resolvable = "const"
)

type ResolveError struct {
	Value string
	Kind  Resolvable
	Node  ast.Node
}

func (e *ResolveError) Error() string {
	details := ""
	if e.Kind == TAG {
		details = ", you did not specify '" + e.Value + "' as a target"
	}

	return Unresolved + " " + string(e.Kind) + ": " + e.Value + details
}

func (e *ResolveError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type TypeError struct {
	Message string
	Node    ast.Node
}

func (e *TypeError) Error() string {
	return Type + ": " + e.Message
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
	start, end := e.Original.Start(), e.Original.End()
	pos := start.String() + " - " + end.String()

	return Duplicate + ": '" + e.Name + "' is already defined here " + pos
}

func (e *DuplicateDefError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type TargetMismatchError struct {
	Target  string
	Tag     language.Tag
	Missing bool
	Node    ast.Node
}

func (e *TargetMismatchError) Error() string {
	if e.Missing {
		return TargetMismatch + ": key is missing the '" + e.Tag.String() + "' field"
	}

	return TargetMismatch
}

func (e *TargetMismatchError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}
