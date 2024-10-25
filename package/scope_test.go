package pkg_test

import (
	"testing"

	"github.com/CanPacis/lcl/errs"
	pkg "github.com/CanPacis/lcl/package"
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/test"
	"github.com/CanPacis/lcl/types"
	"github.com/stretchr/testify/assert"
)

type ScopeDep struct {
	scope *pkg.Scope
}

func (c *ScopeDep) Inject(dep *pkg.Scope) {
	c.scope = dep
}

type RegisterCase struct {
	ScopeDep

	In  *ast.FnDefStmt
	Err error
}

func (c *RegisterCase) Run(assert *assert.Assertions) {
	err := c.scope.RegisterFn(c.In)
	if c.Err == nil {
		assert.NoError(err)
	} else {
		assert.ErrorIs(err, c.Err)
	}
}

type ResolveCase struct {
	ScopeDep

	In  ast.Expr
	Out types.Type
	Err error
}

func (c *ResolveCase) Run(assert *assert.Assertions) {
	typ, err := c.scope.ResolveExpr(c.In)
	assert.Equal(c.Out, typ)
	if c.Err == nil {
		assert.NoError(err)
	} else {
		assert.ErrorIs(err, c.Err)
	}
}

func TestScope(t *testing.T) {
	scope := pkg.NewScope()

	duplicate := &ast.FnDefStmt{
		Name:   &ast.IdentExpr{Value: "Duplicate"},
		Params: nil,
		Body:   nil,
	}
	scope.RegisterFn(duplicate)

	scope.DefineBuiltin("itoa", &types.Fn{
		In:  []types.Type{types.Int},
		Out: types.String,
	})

	scope.DefineBuiltin("bti", &types.Fn{
		In:  []types.Type{types.Bool},
		Out: types.Int,
	})

	scope.Define("newstr", types.New("str", types.NewList(types.Rune)))

	scope.Define("user",
		types.NewStruct(
			types.NewPair(0, "name", types.String),
			types.NewPair(1, "age", types.U8),
		),
	)

	tests := []test.Injector[*pkg.Scope]{
		&RegisterCase{
			In:  duplicate,
			Err: errs.ErrDuplicateDefinition,
		},
		&RegisterCase{
			In: &ast.FnDefStmt{
				Name:   &ast.IdentExpr{Value: "Undefined"},
				Params: []*ast.Parameter{},
				Body:   &ast.IdentExpr{Value: "undefined"},
			},
		},
		&RegisterCase{
			In: &ast.FnDefStmt{
				Name: &ast.IdentExpr{Value: "DuplicateParams"},
				Params: []*ast.Parameter{
					{
						Index: 0,
						Name:  &ast.IdentExpr{Value: "dup"},
						Type:  &ast.IdentExpr{Value: "int"},
					},
					{
						Index: 1,
						Name:  &ast.IdentExpr{Value: "dup"},
						Type:  &ast.IdentExpr{Value: "string"},
					},
				},
				Body: &ast.EmptyExpr{},
			},
			Err: errs.ErrDuplicateDefinition,
		},
		&ResolveCase{
			In:  &ast.IdentExpr{Value: "undefined"},
			Out: types.Invalid,
			Err: errs.ErrUnresolvedConstReference,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "bti"},
				Args: []ast.Expr{
					&ast.IdentExpr{Value: "false"},
				},
			},
			Out: types.Int,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "bti"},
				Args: []ast.Expr{
					&ast.IdentExpr{Value: "undefined"},
				},
			},
			Out: types.Int,
			Err: errs.ErrUnresolvedConstReference,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "bti"},
				Args: []ast.Expr{
					&ast.IdentExpr{Value: "true"},
					&ast.IdentExpr{Value: "false"},
				},
			},
			Out: types.Int,
			Err: errs.ErrTooManyArguments,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn:   &ast.IdentExpr{Value: "bti"},
				Args: []ast.Expr{},
			},
			Out: types.Int,
			Err: errs.ErrTooFewArguments,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "itoa"},
				Args: []ast.Expr{
					&ast.CallExpr{
						Fn: &ast.IdentExpr{Value: "bti"},
						Args: []ast.Expr{
							&ast.IdentExpr{Value: "true"},
						},
					},
				},
			},
			Out: types.String,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn:   &ast.IdentExpr{Value: "true"},
				Args: []ast.Expr{},
			},
			Out: types.Invalid,
			Err: errs.ErrNotCallable,
		},
		&ResolveCase{
			In: &ast.MemberExpr{
				Left:  &ast.IdentExpr{Value: "true"},
				Right: &ast.IdentExpr{Value: "member"},
			},
			Out: types.Invalid,
			Err: errs.ErrNotIndexable,
		},
		&ResolveCase{
			In: &ast.IndexExpr{
				Host:  &ast.IdentExpr{Value: "true"},
				Index: &ast.IdentExpr{Value: "member"},
			},
			Out: types.Invalid,
			Err: errs.ErrNotIndexable,
		},
		&ResolveCase{
			In: &ast.ImportExpr{
				Left:  &ast.IdentExpr{Value: "unresolved"},
				Right: &ast.IdentExpr{Value: "import"},
			},
			Out: types.Invalid,
			Err: errs.ErrUnresolvedImportReference,
		},
		&ResolveCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "unresolved"},
			},
			Out: types.Invalid,
			Err: errs.ErrUnresolvedFnReference,
		},
		&ResolveCase{
			In: &ast.BinaryExpr{
				Left:  &ast.IdentExpr{Value: "true"},
				Right: &ast.NumberLitExpr{Value: 0},
			},
			Out: types.Bool,
			Err: errs.ErrNotComparable,
		},
		&ResolveCase{
			In: &ast.BinaryExpr{
				Left:  &ast.IdentExpr{Value: "newstr"},
				Right: &ast.StringLitExpr{Value: ""},
			},
			Out: types.Bool,
		},
		&ResolveCase{
			In: &ast.BinaryExpr{
				Left:  &ast.NumberLitExpr{Value: 5},
				Right: &ast.IdentExpr{Value: "newstr"},
			},
			Out: types.Bool,
			Err: errs.ErrNotComparable,
		},
		&ResolveCase{
			In: &ast.MemberExpr{
				Left:  &ast.IdentExpr{Value: "user"},
				Right: &ast.IdentExpr{Value: "name"},
			},
			Out: types.String,
		},
		&ResolveCase{
			In: &ast.MemberExpr{
				Left:  &ast.IdentExpr{Value: "user"},
				Right: &ast.IdentExpr{Value: "age"},
			},
			Out: types.U8,
		},
		&ResolveCase{
			In: &ast.MemberExpr{
				Left:  &ast.IdentExpr{Value: "user"},
				Right: &ast.IdentExpr{Value: "invalid"},
			},
			Out: types.Invalid,
			Err: errs.ErrInvalidIndex,
		},
		&ResolveCase{
			In: &ast.IndexExpr{
				Host:  &ast.StringLitExpr{Value: "string"},
				Index: &ast.NumberLitExpr{Value: 0},
			},
			Out: types.Rune,
		},
		&ResolveCase{
			In: &ast.IndexExpr{
				Host:  &ast.StringLitExpr{Value: "string"},
				Index: &ast.IdentExpr{Value: "true"},
			},
			Out: types.Invalid,
			Err: errs.ErrInvalidIndex,
		},
	}
	test.RunWith(t, tests, scope)
}
