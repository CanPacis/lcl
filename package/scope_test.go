package pkg_test

import (
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/test"
	"github.com/CanPacis/go-i18n/types"
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
	}
	test.RunWith(t, tests, scope)
}
