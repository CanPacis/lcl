package analyzer

import (
	"fmt"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/internal"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
	"golang.org/x/text/language"
)

// type ResolveContext int

// const (
// 	GLOBAL ResolveContext = iota
// 	FN_BODY
// 	TEMPLATE_EXPR

// 	FN
// 	TYPE
// 	CONST
// 	MEMBER
// )

type Section struct {
	Name      string
	Keys      map[string]*Key
	Templates map[string]*Template
	Sections  []*Section
}

type Template struct {
	Name   string
	Params []types.Type
	// TODO: figure this out?
	Fields map[language.Tag]int
}

type Key struct {
	Name   string
	Fields map[language.Tag]string
}

type Checker struct {
	env *types.Environment

	tags map[string]language.Tag

	types   map[string]*ast.TypeDefStmt
	fns     map[string]*ast.FnDefStmt
	targets map[string]*ast.DeclTarget

	ctx   *internal.Stack[errs.Resolvable]
	scope *internal.Stack[*pkg.Scope]
}

func (c *Checker) ResolveType(expr ast.TypeExpr) (types.Type, error) {
	c.ctx.Push(errs.TYPE)
	defer c.ctx.Pop()

	err := &errs.ResolveError{
		Value: "unknown",
		Kind:  errs.TYPE,
		Node:  expr,
	}

	switch expr := expr.(type) {
	case *ast.IdentExpr:
		typ, ok := c.env.Lookup(expr.Value)
		if ok {
			return typ, nil
		}

		err.Value = expr.Value
	case *ast.MemberExpr:
		// c.env.Lookup(expr.Left)
		// typ, ok := c.env.Lookup(expr.Right.Value, expr.Left.Value)
		// if ok {
		// 	return typ, nil
		// }

		// err.Value = expr.Left.Value + "." + expr.Right.Value
		// TODO: implement
		panic("not implemented")
	case *ast.StructLitExpr:
		pairs := []types.TypePair{}

		for _, field := range expr.Fields {
			typ, err := c.ResolveType(field.Type)
			if err != nil {
				return types.Empty, err
			}

			pairs = append(pairs, types.NewPair(field.Index, field.Name.Value, typ))
		}

		return types.NewStruct(pairs...), nil
	case *ast.ListTypeExpr:
		typ, err := c.ResolveType(expr.Type)
		if err != nil {
			return types.Empty, err
		}

		return types.NewList(typ), nil
	}

	return types.Empty, err
}

func (c *Checker) ResolveExpr(expr ast.Expr) (types.Type, error) {
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		left, err := c.ResolveExpr(expr.Left)
		if err != nil {
			return types.Empty, err
		}

		right, err := c.ResolveExpr(expr.Right)
		if err != nil {
			return types.Empty, err
		}

		if !c.Comparable(left, right) {
			return types.Empty, errs.NewTypeError(expr, "%s: %s %s", errs.NotComparable, left.Name(), right.Name())
		}
		return types.Bool, nil
	case *ast.TernaryExpr:
		pred, err := c.ResolveExpr(expr.Predicate)
		if err != nil {
			return types.Empty, err
		}

		left, err := c.ResolveExpr(expr.Left)
		if err != nil {
			return types.Empty, err
		}

		right, err := c.ResolveExpr(expr.Right)
		if err != nil {
			return types.Empty, err
		}

		if pred != types.Bool {
			return types.Empty, errs.NewTypeError(expr, "%s, got %s", errs.PredIsNonBool, pred.Name())
		}

		if !c.Convertible(left, right) {
			return types.Empty, errs.NewTypeError(expr, "%s, got %s and %s", errs.PredIsInvalid, left.Name(), right.Name())
		}

		return left, nil
	case *ast.CallExpr:
		c.ctx.Push(errs.FN)
		fn, err := c.ResolveExpr(expr.Fn)
		if err != nil {
			return types.Empty, err
		}
		c.ctx.Pop()

		callable, ok := fn.(*types.Fn)
		if !ok {
			return types.Empty, errs.NewTypeError(expr.Fn, errs.NotCallable)
		}

		if len(callable.In) != len(expr.Args) {
			return callable.Out, errs.NewTypeError(
				expr.Fn,
				"%s, fn expects %d params but %d is given",
				errs.ArgumentCount,
				len(callable.In),
				len(expr.Args),
			)
		}

		for i, arg := range expr.Args {
			typ, err := c.ResolveExpr(arg)
			if err != nil {
				return callable.Out, err
			}
			param := callable.In[i]

			if !c.Assignable(param, typ) {
				return types.Empty, errs.NewTypeError(
					arg,
					"%s, fn expects a '%s' here but '%s' is given",
					errs.NotAssignable,
					param.Name(),
					typ.Name(),
				)
			}
		}

		return callable.Out, nil
	case *ast.MemberExpr:
		c.ctx.Push(errs.CONST)
		defer c.ctx.Pop()

		typ, err := c.ResolveExpr(expr.Left)
		if err != nil {
			return types.Empty, err
		}
		fmt.Println("MEMBER", typ)
		// TODO: implement?
		panic("not implemented")
	case *ast.IndexExpr:
		// TODO: implement?
		panic("not implemented")
	case *ast.GroupExpr:
		return c.ResolveExpr(expr.Expr)
	case *ast.IdentExpr:
		typ, ok := c.scope.Last().Lookup(expr.Value)
		if ok {
			return typ, nil
		}

		return types.Empty, &errs.ResolveError{
			Value: expr.Value,
			Kind:  c.ctx.Last(),
			Node:  expr,
		}
	case *ast.StringLitExpr:
		return types.String, nil
	case *ast.TemplateLitExpr:
		for _, part := range expr.Value {
			typ, err := c.ResolveExpr(part)
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
		return types.Float, nil
	default:
		return types.Empty, nil

	}
}

func (c *Checker) Comparable(left, right types.Type) bool {
	return left.Name() == right.Name()
}

func (c *Checker) Convertible(left, right types.Type) bool {
	return left.Name() == right.Name()
}

func (c *Checker) Assignable(left, right types.Type) bool {
	return left.Name() == right.Name()
}

func (c *Checker) RegisterType(node *ast.TypeDefStmt) error {
	if original, exists := c.types[node.Name.Value]; exists {
		return &errs.DuplicateError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	c.types[node.Name.Value] = node
	c.env.Define(node.Name.Value, types.Empty)
	return nil
}

func (c *Checker) RegisterFn(node *ast.FnDefStmt) error {
	if original, exists := c.fns[node.Name.Value]; exists {
		return &errs.DuplicateError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	c.fns[node.Name.Value] = node
	c.scope.Last().Define(node.Name.Value, types.Empty)
	return nil
}

func (c *Checker) RegisterTarget(node *ast.DeclTarget) error {
	if original, exists := c.fns[node.Name.Value]; exists {
		return &errs.DuplicateError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	var name string
	if node.Tag != nil {
		name = node.Tag.Value
	} else {
		name = node.Name.Value
	}

	c.targets[node.Name.Value] = node
	tag, err := language.Parse(name)
	if err != nil {
		return &errs.ResolveError{
			Value: name,
			Kind:  errs.TARGET,
			Node:  node,
		}
	}
	c.tags[node.Name.Value] = tag
	return nil
}

func (c *Checker) LookupTag(expr *ast.IdentExpr) (language.Tag, error) {
	for name, tag := range c.tags {
		if name == expr.Value {
			return tag, nil
		}
	}
	return language.Tag{}, &errs.ResolveError{
		Value: expr.Value,
		Kind:  errs.TAG,
		Node:  expr,
	}
}

func NewChecker(scope *pkg.Scope, env *types.Environment) *Checker {
	c := &Checker{
		env: env,

		tags: make(map[string]language.Tag),

		ctx:   internal.NewStack(errs.CONST),
		scope: internal.NewStack(scope),

		types:   make(map[string]*ast.TypeDefStmt),
		fns:     make(map[string]*ast.FnDefStmt),
		targets: make(map[string]*ast.DeclTarget),
	}
	return c
}
