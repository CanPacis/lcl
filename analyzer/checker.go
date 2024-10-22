package analyzer

import (
	"errors"
	"fmt"

	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
)

type ResolveContext int

const (
	GLOBAL ResolveContext = iota
	PROC_BODY
	TEMPLATE_EXPR
)

type Checker struct {
	env   *types.Environment
	scope *pkg.Scope

	frame []ResolveContext
	init  bool
}

func (c *Checker) Init() {
	if c.init {
		return
	}
	c.frame = append(c.frame, GLOBAL)
	c.init = true
}

func (c *Checker) Begin(ctx ResolveContext) {
	c.frame = append(c.frame, ctx)
}

func (c *Checker) End() {
	c.frame = c.frame[:len(c.frame)-1]
}

func (c Checker) ContextFrame() ResolveContext {
	return c.frame[len(c.frame)-1]
}

func (c *Checker) ResolveType(expr ast.TypeExpr) (types.Type, error) {
	err := &errs.ResolveError{
		Value: "unknown",
		Kind:  errs.TYPE,
		Node:  expr,
	}

	switch expr := expr.(type) {
	case *ast.IdentExpr:
		typ, ok := c.env.Lookup(expr.Value, "")
		if ok {
			return typ, nil
		}

		err.Value = expr.Value
	case *ast.TypeMemberExpr:
		typ, ok := c.env.Lookup(expr.Right.Value, expr.Left.Value)
		if ok {
			return typ, nil
		}

		err.Value = expr.Left.Value + "." + expr.Right.Value
	case *ast.StructLitExpr:
		fields := map[string]types.Type{}

		for _, field := range expr.List {
			typ, err := c.ResolveType(field.Type)
			if err != nil {
				return types.Empty, err
			}
			fields[field.Name.Value] = typ
		}

		return types.NewStruct(fields), nil
	case *ast.ListTypeExpr:
		typ, err := c.ResolveType(expr.Type)
		if err != nil {
			return types.Empty, err
		}

		return &types.List{Type: typ}, nil
	default:
		return types.Empty, errors.New("???")
	}

	return types.Empty, err
}

func (c *Checker) ResolveExpr(expr ast.Expr) (types.Type, error) {
	err := &errs.ResolveError{
		Value: "unknown",
		Kind:  errs.CONST,
		Node:  expr,
	}

	switch expr := expr.(type) {
	case *ast.SelfExpr:
		return types.Self, nil
	case *ast.BinaryExpr:
		fmt.Println(expr.Left, expr.Right)
		return types.Bool, nil
	case *ast.TernaryExpr:
		fmt.Println(expr.Predicate, expr.Left, expr.Right)
		return types.Bool, nil
	case *ast.ProcCallExpr:
		return types.Empty, errors.ErrUnsupported
	case *ast.MemberExpr:
		return types.Empty, errors.ErrUnsupported
	case *ast.IndexExpr:
		return types.Empty, errors.ErrUnsupported
	case *ast.GroupExpr:
		return c.ResolveExpr(expr.Expr)
	case *ast.IdentExpr:
		typ, ok := c.scope.Lookup(expr.Value, "")
		if ok {
			return typ, nil
		}

		err.Value = expr.Value
	case *ast.StringLitExpr:
		return types.String, nil
	case *ast.TemplateLitExpr:
		// TODO: create a new type for templates?
		return types.String, nil
	case *ast.NumberLitExpr:
		return types.Int, nil
	case *ast.EmptyExpr:
		return types.Empty, nil
	default:
		return types.Empty, errors.New("???")
	}

	return types.Empty, err
}

func (c *Checker) Assignable(left, right types.Type) error {
	if left == types.Self {
		return nil
	}
	fmt.Println(left.Name(), right.Name())
	return errors.ErrUnsupported
}

func NewChecker(scope *pkg.Scope, env *types.Environment) *Checker {
	return &Checker{
		scope: scope,
		env:   env,
	}
}
