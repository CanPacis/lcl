package errs

import (
	"errors"
	"fmt"

	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/parser/token"
)

type Type interface {
	String() string
}

var (
	// Type errors

	ErrInvalidType      = errors.New("invalid type")
	ErrCannotUseType    = errors.New("cannot use type")
	ErrNotComparable    = errors.New("expressions are not comparable")
	ErrNotOperable      = errors.New("expressions are not operable")
	ErrNotCallable      = errors.New("expression is not callable")
	ErrNotIndexable     = errors.New("expression is not indexable")
	ErrInvalidIndex     = errors.New("invalid index")
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
	Value string
}

func (e *TypeError) Error() string {
	switch {
	case errors.Is(e.Err, ErrInvalidType):
		return fmt.Sprintf("%s: %s, expected %s but got %s", e.Name(), e.Err.Error(), e.Left.String(), e.Right.String())
	case errors.Is(e.Err, ErrCannotUseType):
		return fmt.Sprintf("%s: %s %s as %s", e.Name(), e.Err.Error(), e.Left.String(), e.Right.String())
	case errors.Is(e.Err, ErrNotComparable), errors.Is(e.Err, ErrNotOperable), errors.Is(e.Err, ErrMultipleTypes):
		return fmt.Sprintf("%s: %s, %s != %s", e.Name(), e.Err.Error(), e.Left.String(), e.Right.String())
	case errors.Is(e.Err, ErrNotCallable):
		return fmt.Sprintf("%s: %s, %s is not a function", e.Name(), e.Err.Error(), e.Type.String())
	case errors.Is(e.Err, ErrNotIndexable):
		return fmt.Sprintf("%s: %s, %s is not a list or a struct", e.Name(), e.Err.Error(), e.Type.String())
	case errors.Is(e.Err, ErrInvalidIndex):
		return fmt.Sprintf("%s: %s, cannot index %s with a %s", e.Name(), e.Err.Error(), e.Type.String(), e.Value)
	case errors.Is(e.Err, ErrTooManyArguments), errors.Is(e.Err, ErrTooFewArguments):
		return fmt.Sprintf("%s: %s, %s expects %d arguments but got %d", e.Name(), e.Err.Error(), e.Type.String(), e.N, e.M)
	case errors.Is(e.Err, ErrNonBoolPredicate):
		return fmt.Sprintf("%s: %s, this expression should be a bool not a %s", e.Name(), e.Err.Error(), e.Type.String())
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

func (e *TypeError) Ranger() token.Range {
	return e.Node.Range()
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
		return fmt.Sprintf("%s: %s, %s is already defined here %s", e.Name(), e.Err.Error(), e.Value, e.Original.Range())
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

func (e *ReferenceError) Range() token.Range {
	return e.Node.Range()
}
