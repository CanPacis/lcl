package errs

import (
	"errors"
	"fmt"

	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
)

type Type interface {
	String() string
}

var (
	// Type errors

	ErrInvalidType      = errors.New("invalid type")
	ErrCannotUseType    = errors.New("cannot use type")
	ErrNotComparable    = errors.New("expressions are not comparable")
	ErrNotCallable      = errors.New("expression is not callable")
	ErrNotIndexable     = errors.New("expression is not indexable")
	ErrTooManyArguments = errors.New("too many arguments in call")
	ErrTooFewArguments  = errors.New("too few arguments in call")
	ErrNonBoolPredicate = errors.New("non bool predicate")
	ErrMultipleTypes    = errors.New("both sides of this expression must be the same type")
	ErrBuiltinOverride  = errors.New("is a builtin type you cannot override")

	// Reference errors

	ErrInvalidDeclName           = errors.New("invalid declaration name")
	ErrDuplicateDefinition       = errors.New("duplicate definition")
	ErrInvalidTargetTag          = errors.New("invalid target tag")
	ErrUndeclaredTargetTag       = errors.New("undeclared target tag")
	ErrMissingTargetField        = errors.New("missing target field")
	ErrUnresolvedImportReference = errors.New("unresolved import reference")
	ErrUnresolvedTypeReference   = errors.New("unresolved type reference")
	ErrUnresolvedFnReference     = errors.New("unresolved fn reference")
	ErrUnresolvedConstReference  = errors.New("unresolved const reference")
)

type TypeError struct {
	Err   error
	Node  ast.Node
	Left  Type
	Right Type
	Type  Type
	N     int
	M     int
}

func (e *TypeError) Error() string {
	switch {
	case errors.Is(e.Err, ErrInvalidType):
		return fmt.Sprintf("%s: %s", e.Name(), e.Err.Error())
	case errors.Is(e.Err, ErrCannotUseType):
		return fmt.Sprintf("%s: %s %s as %s", e.Name(), e.Err.Error(), e.Left.String(), e.Right.String())
	case errors.Is(e.Err, ErrNotComparable):
		return fmt.Sprintf("%s: %s, %s != %s", e.Name(), e.Err.Error(), e.Left.String(), e.Right.String())
	case errors.Is(e.Err, ErrNotCallable):
		return fmt.Sprintf("%s: %s, %s is not a function", e.Name(), e.Err.Error(), e.Type.String())
	case errors.Is(e.Err, ErrNotIndexable):
		return fmt.Sprintf("%s: %s, %s is not a list", e.Name(), e.Err.Error(), e.Type.String())
	case errors.Is(e.Err, ErrTooManyArguments), errors.Is(e.Err, ErrTooFewArguments):
		return fmt.Sprintf("%s: %s, %s expects %d arguments but got %d", e.Name(), e.Err.Error(), e.Type.String(), e.N, e.M)
	case errors.Is(e.Err, ErrNonBoolPredicate):
		return fmt.Sprintf("%s: %s, this expression should be a bool", e.Name(), e.Err.Error())
	case errors.Is(e.Err, ErrBuiltinOverride):
		return fmt.Sprintf("%s: %s %s", e.Name(), e.Type.String(), e.Err.Error())
	default:
		return fmt.Sprintf("%s: %s", e.Name(), e.Err.Error())
	}
}

func (e *TypeError) Unwrap() error {
	return e.Err
}

func (e *TypeError) Name() string {
	return "type error"
}

func (e *TypeError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}

type ReferenceError struct {
	Err      error
	Node     ast.Node
	Original ast.Node
	Value    string
}

func (e *ReferenceError) Error() string {
	switch {
	case errors.Is(e.Err, ErrInvalidDeclName):
		return fmt.Sprintf("%s: %s '%s'", e.Name(), e.Err.Error(), e.Value)
	case errors.Is(e.Err, ErrDuplicateDefinition):
		pos := fmt.Sprintf("%s - %s", e.Original.Start(), e.Original.End())
		return fmt.Sprintf("%s: %s, %s is already defined here %s", e.Name(), e.Err.Error(), e.Value, pos)
	case errors.Is(e.Err, ErrInvalidTargetTag):
		return fmt.Sprintf("%s: %s '%s'", e.Name(), e.Err.Error(), e.Value)
	case errors.Is(e.Err, ErrUndeclaredTargetTag):
		return fmt.Sprintf("%s: %s '%s'", e.Name(), e.Err.Error(), e.Value)
	case errors.Is(e.Err, ErrMissingTargetField):
		return fmt.Sprintf("%s: %s, key does not include a field for '%s'", e.Name(), e.Err.Error(), e.Value)
	case errors.Is(e.Err, ErrUnresolvedImportReference),
		errors.Is(e.Err, ErrUnresolvedTypeReference),
		errors.Is(e.Err, ErrUnresolvedFnReference),
		errors.Is(e.Err, ErrUnresolvedConstReference):
		return fmt.Sprintf("%s: %s '%s'", e.Name(), e.Err.Error(), e.Value)
	default:
		return fmt.Sprintf("%s: %s", e.Name(), e.Err.Error())
	}
}

func (e *ReferenceError) Unwrap() error {
	return e.Err
}

func (e *ReferenceError) Name() string {
	return "reference error"
}

func (e *ReferenceError) Position() (start token.Position, end token.Position) {
	return e.Node.Start(), e.Node.End()
}
