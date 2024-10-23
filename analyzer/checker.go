package analyzer

import (
	"errors"

	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
	"golang.org/x/text/language"
)

type ResolveContext int

const (
	GLOBAL ResolveContext = iota
	FN_BODY
	TEMPLATE_EXPR

	FN
	TYPE
	CONST
)

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
	env    *types.Environment
	scopes []*pkg.Scope

	tags map[string]language.Tag

	types   map[string]*ast.TypeDefStmt
	fns     map[string]*ast.FnDefStmt
	targets map[string]*ast.DeclTarget

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

func (c *Checker) BeginCtx(ctx ResolveContext) {
	c.frame = append(c.frame, ctx)
}

func (c *Checker) EndCtx() {
	c.frame = c.frame[:len(c.frame)-1]
}

func (c Checker) Context() ResolveContext {
	return c.frame[len(c.frame)-1]
}

func (c Checker) Scope() *pkg.Scope {
	return c.scopes[len(c.scopes)-1]
}

func (c *Checker) PushScope() {
	c.scopes = append(c.scopes, pkg.NewSubScope(c.Scope()))
}

func (c *Checker) PopScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) ResolveType(expr ast.TypeExpr) (types.Type, error) {
	c.BeginCtx(TYPE)
	defer c.EndCtx()

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
		pairs := []types.TypePair{}

		for _, field := range expr.List {
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
	default:
		return types.Empty, errors.New("???")
	}

	return types.Empty, err
}

func (c *Checker) ResolveExpr(expr ast.Expr) (types.Type, error) {
	var kind errs.Resolvable

	switch c.Context() {
	case FN:
		kind = errs.FN
	case CONST:
		kind = errs.CONST
	default:
		kind = errs.CONST
	}

	err := &errs.ResolveError{
		Value: "unknown",
		Kind:  kind,
		Node:  expr,
	}

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
		c.BeginCtx(FN)
		fn, err := c.ResolveExpr(expr.Fn)
		if err != nil {
			return types.Empty, err
		}
		c.EndCtx()

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
		// TODO: implement?
		panic("not implemented")
	case *ast.IndexExpr:
		// TODO: implement?
		panic("not implemented")
	case *ast.GroupExpr:
		return c.ResolveExpr(expr.Expr)
	case *ast.IdentExpr:
		typ, ok := c.Scope().Lookup(expr.Value)
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
		isInt := expr.Value == float64(int(expr.Value))
		if isInt {
			return types.Int, nil
		}
		return types.Float, nil
	case *ast.EmptyExpr:
		return types.Empty, nil
	}

	return types.Empty, err
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
		return &errs.DuplicateDefError{
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
		return &errs.DuplicateDefError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	c.fns[node.Name.Value] = node
	c.Scope().Define(node.Name.Value, types.Empty)
	return nil
}

func (c *Checker) RegisterTarget(node *ast.DeclTarget) (language.Tag, error) {
	if original, exists := c.fns[node.Name.Value]; exists {
		return language.Tag{}, &errs.DuplicateDefError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	c.targets[node.Name.Value] = node
	tag, err := language.Parse(node.Tag.Value)
	if err != nil {
		return language.Tag{}, &errs.ResolveError{
			Value: node.Tag.Value,
			Kind:  errs.TARGET,
			Node:  node,
		}
	}
	c.tags[node.Name.Value] = tag
	return tag, nil
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
		scopes: []*pkg.Scope{scope},
		env:    env,

		tags: make(map[string]language.Tag),

		types:   make(map[string]*ast.TypeDefStmt),
		fns:     make(map[string]*ast.FnDefStmt),
		targets: make(map[string]*ast.DeclTarget),
	}
	c.Init()
	return c
}
