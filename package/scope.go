package pkg

import (
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/internal"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
	"github.com/CanPacis/go-i18n/types"
)

type Context int

const (
	CONST Context = iota
	IMPORT
	FN
)

type Scope struct {
	imports map[string]*Scope
	builtin map[string]types.Type
	objects map[string]types.Type

	importDefs map[string]*ast.IdentExpr
	fnDefs     map[string]*ast.FnDefStmt

	ctx    *internal.Stack[Context]
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

func (s *Scope) Pop() *Scope {
	return s.parent
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

func (s *Scope) DefineBuiltin(name string, typ types.Type) {
	s.builtin[name] = typ
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

		if !left.Comparable(right) {
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

		var op types.Operation
		switch expr.Operator.Kind {
		case token.PLUS:
			op = types.Addition
		case token.MINUS:
			op = types.Subtraction
		case token.STAR:
			op = types.Multiplication
		case token.FORWARD_SLASH:
			op = types.Division
		}

		if !left.Operable(right, op) {
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

		if !left.Convertible(right) {
			return types.Invalid, &errs.TypeError{
				Err:   errs.ErrMultipleTypes,
				Node:  expr,
				Left:  left,
				Right: right,
			}
		}

		return left, nil
	case *ast.CallExpr:
		s.ctx.Push(FN)
		defer s.ctx.Pop()
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
			s.ctx.Push(CONST)
			typ, err := s.ResolveExpr(arg)
			s.ctx.Pop()
			if err != nil {
				return callable.Out, err
			}
			param := callable.In[i]

			if !param.Assignable(typ) {
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

		typ, ok := indexer.Index(expr.Right.Value)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:   errs.ErrInvalidIndex,
				Type:  left,
				Node:  expr.Right,
				Value: expr.Right.Value,
			}
		}
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
		host, err := s.ResolveExpr(expr.Host)
		if err != nil {
			return types.Invalid, err
		}
		indexer, ok := host.(types.Indexer)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:  errs.ErrNotIndexable,
				Type: host,
				Node: expr.Host,
			}
		}

		index, err := s.ResolveExpr(expr.Index)
		if err != nil {
			return types.Invalid, err
		}

		if index != types.Int {
			return types.Invalid, &errs.TypeError{
				Err:   errs.ErrInvalidIndex,
				Type:  host,
				Node:  expr.Index,
				Value: index.String(),
			}
		}

		typ, ok := indexer.Index(0)
		if !ok {
			return types.Invalid, &errs.TypeError{
				Err:   errs.ErrInvalidIndex,
				Type:  host,
				Node:  expr.Host,
				Value: index.String(),
			}
		}
		return typ, nil
	case *ast.GroupExpr:
		return s.ResolveExpr(expr.Expr)
	case *ast.IdentExpr:
		typ, ok := s.lookup(expr.Value)
		if !ok {
			var e error

			switch s.ctx.Last() {
			case CONST:
				e = errs.ErrUnresolvedConstReference
			case FN:
				e = errs.ErrUnresolvedFnReference
			case IMPORT:
				e = errs.ErrUnresolvedImportReference
			default:
				e = errs.ErrUnresolvedConstReference
			}

			return types.Invalid, &errs.ReferenceError{
				Err:   e,
				Value: expr.Value,
				Node:  expr,
			}
		}
		return typ, nil
	case *ast.StringLitExpr:
		return types.String, nil
	case *ast.TemplateLitExpr:
		in := []types.Type{}

		for _, part := range expr.Value {
			typ, err := s.ResolveExpr(part)
			if err != nil {
				return types.NewTemplate([]types.Type{}), err
			}
			in = append(in, typ)
		}

		return types.NewTemplate(in), nil
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

		ctx: internal.NewStack(CONST),

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

		ctx:    internal.NewStack(CONST),
		parent: parent,
	}
}
