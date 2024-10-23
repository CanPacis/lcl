package analyzer_test

import (
	"testing"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/test"
	"github.com/CanPacis/go-i18n/types"
	"github.com/stretchr/testify/assert"
)

type TypeCase struct {
	In  string
	Out types.Type
	Err error

	checker *analyzer.Checker
}

func (c *TypeCase) Inject(dep *analyzer.Checker) {
	c.checker = dep
}

func (c *TypeCase) Run(assert *assert.Assertions) {
	expr := test.MustParseTypeExpr(test.WithSourceString(c.In))
	out, err := c.checker.ResolveType(expr)
	if c.Err != nil {
		assert.ErrorIs(err, c.Err)
	} else {
		assert.NoError(err)
	}
	assert.Equal(c.Out.Name(), out.Name())
}

type ExprCase struct {
	In       string
	Out      types.Type
	Err      error
	Contains string

	checker *analyzer.Checker
}

func (c *ExprCase) Inject(dep *analyzer.Checker) {
	c.checker = dep
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	expr := test.MustParseExpr(test.WithSourceString(c.In))
	out, err := c.checker.ResolveExpr(expr)
	if len(c.Contains) != 0 {
		assert.ErrorContains(err, c.Contains)
	}
	if c.Err != nil {
		assert.ErrorIs(err, c.Err)
	} else {
		assert.NoError(err)
	}
	assert.Equal(c.Out.Name(), out.Name())
}

func TestBuiltinTypes(t *testing.T) {
	var checker = analyzer.NewChecker(pkg.NewScope(), types.NewEnvironment())

	tests := []test.Injector[*analyzer.Checker]{
		&TypeCase{
			In:  "string",
			Out: types.String,
			Err: nil,
		},
		&TypeCase{
			In:  "int",
			Out: types.Int,
			Err: nil,
		},
		&TypeCase{
			In:  "float",
			Out: types.Float,
			Err: nil,
		},
		&TypeCase{
			In:  "bool",
			Out: types.Bool,
			Err: nil,
		},
		&TypeCase{
			In:  "int[]",
			Out: types.NewList(types.Int),
			Err: nil,
		},
		&TypeCase{
			In: "{ name string age int }",
			Out: types.NewStruct(
				types.NewPair(0, "name", types.String),
				types.NewPair(1, "age", types.Int),
			),
			Err: nil,
		},
	}
	test.RunWith(t, tests, checker)
}

func TestBasicExpr(t *testing.T) {
	var checker = analyzer.NewChecker(pkg.NewScope(), types.NewEnvironment())

	tests := []test.Injector[*analyzer.Checker]{
		&ExprCase{
			In:  `"string"`,
			Out: types.String,
			Err: nil,
		},
		&ExprCase{
			In:  "5",
			Out: types.Int,
			Err: nil,
		},
		&ExprCase{
			In:  "5.4",
			Out: types.Float,
			Err: nil,
		},
		&ExprCase{
			In:  "true",
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  "false",
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  "undefined",
			Out: types.Empty,
			Err: &errs.ResolveError{},
		},
	}
	test.RunWith(t, tests, checker)
}

func TestComplexExpr(t *testing.T) {
	scope := pkg.NewScope()
	scope.Define("age", types.Int)
	scope.Define("func", &types.Fn{
		In:  []types.Type{types.Int},
		Out: types.String,
	})

	checker := analyzer.NewChecker(scope, types.NewEnvironment())

	tests := []test.Injector[*analyzer.Checker]{
		&ExprCase{
			In:  `5 == 6`,
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  `(3.1)`,
			Out: types.Float,
			Err: nil,
		},
		&ExprCase{
			In:  `(3 == 3)`,
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:       `5 >= 5.5`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotComparable,
		},
		&ExprCase{
			In:       `age > 18 ? "" : 4`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsInvalid,
		},
		&ExprCase{
			In:       `age ? "" : 4`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsNonBool,
		},
		&ExprCase{
			In:  `age > 18 ? "" : ""`,
			Out: types.String,
			Err: nil,
		},
		&ExprCase{
			In:  `age(0)`,
			Out: types.Empty,
			Err: &errs.TypeError{},
		},
		&ExprCase{
			In:       `func("")`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotAssignable,
		},
		&ExprCase{
			In:  `func(0)`,
			Out: types.String,
			Err: nil,
		},
	}
	test.RunWith(t, tests, checker)
}

func TestResolveErrs(t *testing.T) {
	var checker = analyzer.NewChecker(pkg.NewScope(), types.NewEnvironment())

	tests := []test.Injector[*analyzer.Checker]{
		&ExprCase{
			In:       "undefined",
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.CONST),
		},
		&ExprCase{
			In:       "undefined(.)",
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.FN),
		},
	}
	test.RunWith(t, tests, checker)
}
