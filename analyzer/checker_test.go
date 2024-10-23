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

type CheckerTypeCase struct {
	In  string
	Out types.Type
	Err error

	checker *analyzer.Checker
}

func (c *CheckerTypeCase) Inject(dep *analyzer.Checker) {
	c.checker = dep
}

func (c *CheckerTypeCase) Run(assert *assert.Assertions) {
	expr := test.MustParseTypeExpr(test.WithSourceString(c.In))
	out, err := c.checker.ResolveType(expr)
	if c.Err != nil {
		assert.ErrorIs(err, c.Err)
	} else {
		assert.NoError(err)
	}
	assert.Equal(c.Out.Name(), out.Name())
}

type CheckerExprCase struct {
	In       string
	Out      types.Type
	Err      error
	Contains string

	checker *analyzer.Checker
}

func (c *CheckerExprCase) Inject(dep *analyzer.Checker) {
	c.checker = dep
}

func (c *CheckerExprCase) Run(assert *assert.Assertions) {
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
		&CheckerTypeCase{
			In:  "string",
			Out: types.String,
			Err: nil,
		},
		&CheckerTypeCase{
			In:  "int",
			Out: types.Int,
			Err: nil,
		},
		&CheckerTypeCase{
			In:  "float",
			Out: types.Float,
			Err: nil,
		},
		&CheckerTypeCase{
			In:  "bool",
			Out: types.Bool,
			Err: nil,
		},
		&CheckerTypeCase{
			In:  "int[]",
			Out: types.NewList(types.Int),
			Err: nil,
		},
		&CheckerTypeCase{
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
		&CheckerExprCase{
			In:  `"string"`,
			Out: types.String,
			Err: nil,
		},
		&CheckerExprCase{
			In:  "5",
			Out: types.Int,
			Err: nil,
		},
		&CheckerExprCase{
			In:  "5.4",
			Out: types.Float,
			Err: nil,
		},
		&CheckerExprCase{
			In:  "true",
			Out: types.Bool,
			Err: nil,
		},
		&CheckerExprCase{
			In:  "false",
			Out: types.Bool,
			Err: nil,
		},
		&CheckerExprCase{
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
		&CheckerExprCase{
			In:  `5 == 6`,
			Out: types.Bool,
			Err: nil,
		},
		&CheckerExprCase{
			In:  `(3.1)`,
			Out: types.Float,
			Err: nil,
		},
		&CheckerExprCase{
			In:  `(3 == 3)`,
			Out: types.Bool,
			Err: nil,
		},
		&CheckerExprCase{
			In:       `5 >= 5.5`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotComparable,
		},
		&CheckerExprCase{
			In:       `age > 18 ? "" : 4`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsInvalid,
		},
		&CheckerExprCase{
			In:       `age ? "" : 4`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsNonBool,
		},
		&CheckerExprCase{
			In:  `age > 18 ? "" : ""`,
			Out: types.String,
			Err: nil,
		},
		&CheckerExprCase{
			In:  `age(0)`,
			Out: types.Empty,
			Err: &errs.TypeError{},
		},
		&CheckerExprCase{
			In:       `func("")`,
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotAssignable,
		},
		&CheckerExprCase{
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
		&CheckerExprCase{
			In:       "undefined",
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.CONST),
		},
		&CheckerExprCase{
			In:       "undefined(.)",
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.FN),
		},
	}
	test.RunWith(t, tests, checker)
}
