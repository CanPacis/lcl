package types_test

import (
	"testing"

	"github.com/CanPacis/lcl/errs"
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/test"
	"github.com/CanPacis/lcl/types"
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

type CompareCase struct {
	Left   types.Type
	Right  types.Type
	Result bool
}

func (c *CompareCase) Run(assert *assert.Assertions) {
	assert.Equal(c.Result, c.Left.Comparable(c.Right))
}

func TestCompare(t *testing.T) {
	tests := []test.Runner{
		&CompareCase{
			Left:   types.Bool,
			Right:  types.Bool,
			Result: true,
		},
		&CompareCase{
			Left:   types.I32,
			Right:  types.Int,
			Result: true,
		},
		&CompareCase{
			Right:  types.Int,
			Left:   types.I32,
			Result: true,
		},
		&CompareCase{
			Right:  types.Int,
			Left:   types.I64,
			Result: false,
		},
		&CompareCase{
			Right:  types.String,
			Left:   types.String,
			Result: true,
		},
		&CompareCase{
			Right:  types.String,
			Left:   types.NewList(types.Rune),
			Result: true,
		},
		&CompareCase{
			Right:  types.String,
			Left:   types.NewList(types.Bool),
			Result: false,
		},
		&CompareCase{
			Right:  types.String,
			Left:   types.NewList(types.New("RUNE", types.Rune)),
			Result: true,
		},
		&CompareCase{
			Right:  types.String,
			Left:   types.NewList(types.New("U32", types.U32)),
			Result: true,
		},
	}
	test.Run(t, tests)
}
