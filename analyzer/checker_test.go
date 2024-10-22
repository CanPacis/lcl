package analyzer_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
	"github.com/stretchr/testify/assert"
)

type err struct {
	err error
}

func (e *err) Error() string {
	if e == nil {
		return ""
	}
	return "wrapped: " + e.err.Error()
}

func (e *err) Unwrap() error {
	return e.err
}

func (e *err) Is(target error) bool {
	return reflect.TypeOf(e.err).String() == reflect.TypeOf(target).String()
}

func wrap(e error) error {
	if e == nil {
		return nil
	}
	return &err{err: e}
}

var checker = analyzer.NewChecker(pkg.NewScope(), types.NewEnvironment())

type Runner interface {
	Run(*assert.Assertions)
}

type TypeCase struct {
	In  ast.TypeExpr
	Out types.Type
	Err error
}

func (c *TypeCase) Run(assert *assert.Assertions) {
	out, err := checker.ResolveType(c.In)
	if c.Err != nil {
		assert.ErrorIs(wrap(err), c.Err)
	} else {
		assert.NoError(err)
	}
	assert.Equal(c.Out.Name(), out.Name())
}

type ExprCase struct {
	In       ast.Expr
	Out      types.Type
	Err      error
	Contains string
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	out, err := checker.ResolveExpr(c.In)
	if len(c.Contains) != 0 {
		assert.ErrorContains(err, c.Contains)
	}
	if c.Err != nil {
		assert.ErrorIs(wrap(err), c.Err)
	} else {
		assert.NoError(err)
	}
	assert.Equal(c.Out.Name(), out.Name())
}

func Run(cases []Runner, t *testing.T) {
	assert := assert.New(t)

	for _, c := range cases {
		c.Run(assert)
	}
}

func file(src string) *parser.File {
	return parser.NewFile("test.lcl", bytes.NewBuffer([]byte(src)))
}

func expr(src string) ast.Expr {
	expr, err := parser.ParseExpr(file(src))
	if err != nil {
		panic(err)
	}
	return expr
}

func texpr(src string) ast.TypeExpr {
	expr, err := parser.ParseTypeExpr(file(src))
	if err != nil {
		panic(err)
	}
	return expr
}

func TestBuiltinTypes(t *testing.T) {
	tests := []Runner{
		&TypeCase{
			In:  texpr("string"),
			Out: types.String,
			Err: nil,
		},
		&TypeCase{
			In:  texpr("int"),
			Out: types.Int,
			Err: nil,
		},
		&TypeCase{
			In:  texpr("float"),
			Out: types.Float,
			Err: nil,
		},
		&TypeCase{
			In:  texpr("bool"),
			Out: types.Bool,
			Err: nil,
		},
		&TypeCase{
			In:  texpr("int[]"),
			Out: types.NewList(types.Int),
			Err: nil,
		},
		&TypeCase{
			In: texpr("{ name string age int }"),
			Out: types.NewStruct(
				types.NewPair(0, "name", types.String),
				types.NewPair(1, "age", types.Int),
			),
			Err: nil,
		},
	}
	Run(tests, t)
}

func TestBasicExpr(t *testing.T) {
	tests := []Runner{
		&ExprCase{
			In:  expr(`"string"`),
			Out: types.String,
			Err: nil,
		},
		&ExprCase{
			In:  expr("5"),
			Out: types.Int,
			Err: nil,
		},
		&ExprCase{
			In:  expr("5.4"),
			Out: types.Float,
			Err: nil,
		},
		&ExprCase{
			In:  expr("true"),
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  expr("false"),
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  expr("undefined"),
			Out: types.Empty,
			Err: &errs.ResolveError{},
		},
	}
	Run(tests, t)
}

func TestComplexExpr(t *testing.T) {
	scope := pkg.NewScope()
	scope.Define("age", types.Int)
	scope.Define("func", &types.Proc{
		In:  types.Int,
		Out: types.String,
	})
	checker = analyzer.NewChecker(scope, types.NewEnvironment())

	tests := []Runner{
		&ExprCase{
			In:  expr(`5 == 6`),
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:  expr(`(3.1)`),
			Out: types.Float,
			Err: nil,
		},
		&ExprCase{
			In:  expr(`(3 == 3)`),
			Out: types.Bool,
			Err: nil,
		},
		&ExprCase{
			In:       expr(`5 >= 5.5`),
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotComparable,
		},
		&ExprCase{
			In:       expr(`age > 18 ? "" : 4`),
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsInvalid,
		},
		&ExprCase{
			In:       expr(`age ? "" : 4`),
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.PredIsNonBool,
		},
		&ExprCase{
			In:  expr(`age > 18 ? "" : ""`),
			Out: types.String,
			Err: nil,
		},
		&ExprCase{
			In:  expr(`age(.)`),
			Out: types.Empty,
			Err: &errs.TypeError{},
		},
		&ExprCase{
			In:       expr(`func("")`),
			Out:      types.Empty,
			Err:      &errs.TypeError{},
			Contains: errs.NotAssignable,
		},
		&ExprCase{
			In:  expr(`func(0)`),
			Out: types.String,
			Err: nil,
		},
	}
	Run(tests, t)
}

func TestResolveErrs(t *testing.T) {
	tests := []Runner{
		&ExprCase{
			In:       expr("undefined"),
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.CONST),
		},
		&ExprCase{
			In:       expr("undefined(.)"),
			Out:      types.Empty,
			Err:      &errs.ResolveError{},
			Contains: string(errs.PROC),
		},
	}
	Run(tests, t)
}
