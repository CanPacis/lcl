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
	PROC_BODY
	TEMPLATE_EXPR

	PROC
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
	Type   types.Type
	Fields map[string]int
}

type Key struct {
	Name   string
	Fields map[language.Tag]string
}

type Checker struct {
	env   *types.Environment
	scope *pkg.Scope

	tags map[string]language.Tag

	types   map[string]*ast.TypeDefStmt
	procs   map[string]*ast.FnDefStmt
	targets map[string]*ast.IdentExpr

	self types.Type

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

func (c Checker) Context() ResolveContext {
	return c.frame[len(c.frame)-1]
}

func (c *Checker) ResolveType(expr ast.TypeExpr) (types.Type, error) {
	c.Begin(TYPE)
	defer c.End()

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
	case PROC:
		kind = errs.PROC
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
	case *ast.SelfExpr:
		return types.Self, nil
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
			if (left == types.Self || right == types.Self) && c.Context() == PROC_BODY {
				return types.Empty, errs.NewTypeError(expr, "%s: self expression is ambigious", errs.NotInferrable)
			}
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
		c.Begin(PROC)
		proc, err := c.ResolveExpr(expr.Fn)
		if err != nil {
			return types.Empty, err
		}
		c.End()

		callable, ok := proc.(*types.Proc)
		if !ok {
			return types.Empty, errs.NewTypeError(expr.Fn, errs.NotCallable)
		}

		param, err := c.ResolveExpr(expr.Param)
		if err != nil {
			return types.Empty, err
		}

		if !c.Assignable(callable.In, param) {
			return types.Empty, errs.NewTypeError(
				expr.Param,
				"%s, proc expects a '%s' but got '%s'",
				errs.NotAssignable,
				callable.In.Name(),
				param.Name(),
			)
		}

		return callable.Out, nil
	case *ast.MemberExpr:
		return types.Empty, &errs.UnsupportedError{Node: expr}
	case *ast.IndexExpr:
		return types.Empty, &errs.UnsupportedError{Node: expr}
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
	if c.Context() == PROC_BODY && right == types.Self {
		c.self = left
		return true
	}

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

func (c *Checker) RegisterProc(node *ast.FnDefStmt) error {
	if original, exists := c.procs[node.Name.Value]; exists {
		return &errs.DuplicateDefError{
			Name:     node.Name.Value,
			Original: original,
			Node:     node,
		}
	}

	c.procs[node.Name.Value] = node
	c.scope.Define(node.Name.Value, types.Empty)
	return nil
}

func (c *Checker) RegisterTarget(node *ast.IdentExpr) (language.Tag, error) {
	if original, exists := c.procs[node.Value]; exists {
		return language.Tag{}, &errs.DuplicateDefError{
			Name:     node.Value,
			Original: original,
			Node:     node,
		}
	}

	c.targets[node.Value] = node
	tag, err := language.Parse(node.Value)
	if err != nil {
		return language.Tag{}, &errs.ResolveError{
			Value: node.Value,
			Kind:  errs.TARGET,
			Node:  node,
		}
	}
	c.tags[node.Value] = tag
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
		scope: scope,
		env:   env,
		self:  types.Empty,

		tags: make(map[string]language.Tag),

		types:   make(map[string]*ast.TypeDefStmt),
		procs:   make(map[string]*ast.FnDefStmt),
		targets: make(map[string]*ast.IdentExpr),
	}
	c.Init()
	return c
}
