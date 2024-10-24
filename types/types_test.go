package types_test

import (
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/test"
	"github.com/CanPacis/go-i18n/types"
	"github.com/stretchr/testify/assert"
)

type EnvDep struct {
	env *types.Environment
}

func (c *EnvDep) Inject(dep *types.Environment) {
	c.env = dep
}

type RegisterCase struct {
	EnvDep

	In  *ast.TypeDefStmt
	Err error
}

func (c *RegisterCase) Run(assert *assert.Assertions) {
	err := c.env.RegisterType(c.In)
	if c.Err == nil {
		assert.NoError(err)
	} else {
		assert.ErrorIs(err, c.Err)
	}
}

type ImportRegisterCase struct {
	EnvDep

	In  *ast.IdentExpr
	Err error
}

func (c *ImportRegisterCase) Run(assert *assert.Assertions) {
	err := c.env.RegisterImport(c.In)
	if c.Err == nil {
		assert.NoError(err)
	} else {
		assert.ErrorIs(err, c.Err)
	}
}

type ResolveCase struct {
	EnvDep

	In  ast.TypeExpr
	Out types.Type
	Err error
}

func (c *ResolveCase) Run(assert *assert.Assertions) {
	typ, err := c.env.ResolveType(c.In)
	assert.Equal(c.Out, typ)
	if c.Err == nil {
		assert.NoError(err)
	} else {
		assert.ErrorIs(err, c.Err)
	}
}

func TestEnvironment(t *testing.T) {
	env := types.NewEnvironment()
	time := types.New("Time", types.I64)
	env.Define("Time", time)

	data := types.New("Data", types.NewStruct(
		types.NewPair(0, "name", types.String),
		types.NewPair(1, "age", types.U8),
	))

	user := types.NewEnvironment()
	user.Define("Data", data)
	env.Import("User", user)

	tests := []test.Injector[*types.Environment]{
		&RegisterCase{
			In: &ast.TypeDefStmt{
				Name: &ast.IdentExpr{Value: "Duplicate"},
				Type: &ast.IdentExpr{Value: "bool"},
			},
		},
		&RegisterCase{
			In: &ast.TypeDefStmt{
				Name: &ast.IdentExpr{Value: "Duplicate"},
				Type: &ast.IdentExpr{Value: "bool"},
			},
			Err: errs.ErrDuplicateDefinition,
		},
		&RegisterCase{
			In: &ast.TypeDefStmt{
				Name: &ast.IdentExpr{Value: "string"},
				Type: &ast.IdentExpr{Value: "bool"},
			},
			Err: errs.ErrBuiltinOverride,
		},
		&ResolveCase{
			In:  &ast.IdentExpr{Value: "undefined"},
			Out: types.Invalid,
			Err: errs.ErrUnresolvedTypeReference,
		},
		&ResolveCase{
			In:  &ast.IdentExpr{Value: "string"},
			Out: types.String,
		},
		&ResolveCase{
			In:  &ast.IdentExpr{Value: "Time"},
			Out: time,
		},
		&ResolveCase{
			In: &ast.ImportExpr{
				Left:  &ast.IdentExpr{Value: "User"},
				Right: &ast.IdentExpr{Value: "Data"},
			},
			Out: data,
		},
		&ResolveCase{
			In: &ast.ImportExpr{
				Left:  &ast.IdentExpr{Value: "Invalid"},
				Right: &ast.IdentExpr{Value: "Import"},
			},
			Out: types.Invalid,
			Err: errs.ErrUnresolvedImportReference,
		},
		&ImportRegisterCase{
			In: &ast.IdentExpr{Value: "Duplicate"},
		},
		&ImportRegisterCase{
			In:  &ast.IdentExpr{Value: "Duplicate"},
			Err: errs.ErrDuplicateDefinition,
		},
	}
	test.RunWith(t, tests, env)
}
