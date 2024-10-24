package pkg

import (
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
)

type Scope struct {
	imports map[string]*Scope
	builtin map[string]types.Type
	objects map[string]types.Type

	importDefs map[string]*ast.IdentExpr
	fnDefs     map[string]*ast.FnDefStmt

	parent *Scope
}

func (s *Scope) Exports() map[string]types.Type {
	return s.objects
}

func (s *Scope) RegisterImport(def *ast.IdentExpr) error {
	if original, exists := s.importDefs[def.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Value,
		}
	}

	s.importDefs[def.Value] = def
	return nil
}

func (s *Scope) Import(name string, scope *Scope) {
	s.imports[name] = scope
}

func (s *Scope) RegisterFn(def *ast.FnDefStmt) error {
	if original, exists := s.fnDefs[def.Name.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Name.Value,
		}
	}

	s.fnDefs[def.Name.Value] = def
	return nil
}

func (s *Scope) Define(name string, typ types.Type) {
	s.objects[name] = typ
}

func (s *Scope) Comparable(left, right types.Type) bool {
	// TODO: properly implement
	return left.String() == right.String()
}

func (s *Scope) Convertible(left, right types.Type) bool {
	// TODO: properly implement
	return left.String() == right.String()
}

func (s *Scope) Assignable(left, right types.Type) bool {
	// TODO: properly implement
	return left.String() == right.String()
}

func (s *Scope) Operable(left, right types.Type) bool {
	// TODO: properly implement
	return left == types.Int && right == types.Int
}

func (s Scope) ResolveExpr(expr ast.Expr) (types.Type, error) {
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		left, err := s.ResolveExpr(expr.Left)
		if err != nil {
			return types.Invalid, err
		}

		right, err := s.ResolveExpr(expr.Right)
		if err != nil {
			return types.Bool, err
		}

		if !s.Comparable(left, right) {
			return types.Bool, &errs.TypeError{
				Err:   errs.ErrNotComparable,
				Node:  expr,
				Left:  left,
				Right: right,
			}
		}

		return types.Bool, nil
	case *ast.ArithmeticExpr:
		left, err := s.ResolveExpr(expr.Left)
		if err != nil {
			return types.Invalid, err
		}

		right, err := s.ResolveExpr(expr.Right)
		if err != nil {
			return left, err
		}

		if !s.Operable(left, right) {
			return left, &errs.TypeError{
				Err:   errs.ErrNotOperable,
				Node:  expr,
				Left:  left,
				Right: right,
			}
		}

		return left, nil
	case *ast.TernaryExpr:
		pred, err := s.ResolveExpr(expr.Predicate)
		if err != nil {
			return types.Invalid, err
		}

		left, err := s.ResolveExpr(expr.Left)
		if err != nil {
			return types.Invalid, err
		}

		right, err := s.ResolveExpr(expr.Right)
		if err != nil {
			return left, err
		}

		if pred != types.Bool {
			return left, &errs.TypeError{
				Err:  errs.ErrNonBoolPredicate,
				Type: pred,
				Node: expr,
			}
		}

		if !s.Convertible(left, right) {
			return types.Invalid, &errs.TypeError{
				Err:   errs.ErrMultipleTypes,
				Node:  expr,
				Left:  left,
				Right: right,
			}
		}

		return left, nil
	case *ast.CallExpr:
		fn, err := s.ResolveExpr(expr.Fn)
		if err != nil {
			return types.Invalid, err
		}

		callable, ok := fn.(*types.Fn)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:  errs.ErrNotCallable,
				Node: expr.Fn,
				Type: fn,
			}
		}

		if len(callable.In) != len(expr.Args) {
			var e error

			if len(callable.In) > len(expr.Args) {
				e = errs.ErrTooFewArguments
			} else {
				e = errs.ErrTooManyArguments
			}

			return callable.Out, &errs.TypeError{
				Err:  e,
				Node: expr,
				Type: fn,
				N:    len(callable.In),
				M:    len(expr.Args),
			}
		}

		for i, arg := range expr.Args {
			typ, err := s.ResolveExpr(arg)
			if err != nil {
				return callable.Out, err
			}
			param := callable.In[i]

			if !s.Assignable(param, typ) {
				return callable.Out, &errs.TypeError{
					Err:   errs.ErrInvalidType,
					Left:  param,
					Right: typ,
					Node:  arg,
				}
			}
		}

		return callable.Out, nil
	case *ast.MemberExpr:
		left, err := s.ResolveExpr(expr.Left)
		if err != nil {
			return types.Invalid, err
		}
		indexer, ok := left.(types.Indexer)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:  errs.ErrNotIndexable,
				Type: left,
				Node: expr.Left,
			}
		}

		// TODO: check invalid type
		typ, _ := indexer.Index(expr.Right.Value)
		return typ, nil
	case *ast.ImportExpr:
		scope, ok := s.imports[expr.Left.Value]
		if !ok {
			return types.Invalid, &errs.ReferenceError{
				Err:   errs.ErrUnresolvedImportReference,
				Value: expr.Left.Value,
				Node:  expr.Left,
			}
		}

		return scope.ResolveExpr(expr.Right)
	case *ast.IndexExpr:
		left, err := s.ResolveExpr(expr.Host)
		if err != nil {
			return types.Invalid, err
		}
		indexer, ok := left.(types.Indexer)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:  errs.ErrNotIndexable,
				Type: left,
				Node: expr.Host,
			}
		}

		index, err := s.ResolveExpr(expr.Index)
		if err != nil {
			return types.Invalid, err
		}

		if index != types.Int {
			return types.Invalid, &errs.TypeError{
				Err:  errs.ErrNonIntIndex,
				Type: index,
				Node: expr.Index,
			}
		}

		// TODO: check invalid type
		typ, _ := indexer.Index(0)
		return typ, nil
	case *ast.GroupExpr:
		return s.ResolveExpr(expr.Expr)
	case *ast.IdentExpr:
		typ, ok := s.lookup(expr.Value)
		if !ok {
			return types.Invalid, &errs.ReferenceError{
				Err:   errs.ErrUnresolvedConstReference,
				Value: expr.Value,
				Node:  expr,
			}
		}
		return typ, nil
	case *ast.StringLitExpr:
		return types.String, nil
	case *ast.TemplateLitExpr:
		for _, part := range expr.Value {
			typ, err := s.ResolveExpr(part)
			if err != nil {
				return typ, err
			}
		}

		// TODO
		return types.NewTemplate(nil), nil
	case *ast.NumberLitExpr:
		isInt := expr.Value == float64(int(expr.Value))
		if isInt {
			return types.Int, nil
		}
		return types.F64, nil
	default:
		return types.Invalid, &errs.TypeError{
			Err:   errs.ErrInvalidType,
			Node:  expr,
			Left:  types.Invalid,
			Right: types.Invalid,
		}
	}
}

func (s Scope) lookup(name string) (types.Type, bool) {
	// If parent is not empty, check the local definitions first
	if s.parent != nil {
		typ, ok := s.objects[name]
		if ok {
			return typ, true
		}

		return s.parent.lookup(name)
	}
	// First check the builtins
	typ, ok := s.builtin[name]
	if ok {
		return typ, true
	}

	typ, ok = s.objects[name]
	if ok {
		return typ, true
	}

	return types.Invalid, false
}

func NewScope() *Scope {
	return &Scope{
		imports:    make(map[string]*Scope),
		objects:    make(map[string]types.Type),
		importDefs: make(map[string]*ast.IdentExpr),
		fnDefs:     make(map[string]*ast.FnDefStmt),
		builtin: map[string]types.Type{
			"true":  types.Bool,
			"false": types.Bool,
		},
	}
}

func NewSubScope(parent *Scope) *Scope {
	return &Scope{
		imports:    make(map[string]*Scope),
		objects:    make(map[string]types.Type),
		builtin:    make(map[string]types.Type),
		importDefs: make(map[string]*ast.IdentExpr),
		fnDefs:     make(map[string]*ast.FnDefStmt),

		parent: parent,
	}
}
