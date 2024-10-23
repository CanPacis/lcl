package analyzer_test

import (
	"embed"
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/test"
	"github.com/CanPacis/go-i18n/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

//go:embed test/*.lcl
var tests embed.FS

func semantics(name string, scope *pkg.Scope, env *types.Environment) *analyzer.Semantics {
	source, _ := tests.Open(name)
	defer source.Close()
	file := parser.NewFile(name, source)
	ast := test.MustParse(test.WithFile(file))
	if scope == nil {
		scope = pkg.NewScope()
	}
	if env == nil {
		env = types.NewEnvironment()
	}
	return analyzer.New(file, ast, analyzer.NewChecker(scope, env))
}

func TestDeclare(t *testing.T) {
	assert := assert.New(t)
	s := semantics("test/declare.lcl", nil, nil)

	assert.Equal("i18n", s.ScanName())
	targets := s.ScanTags()

	assert.Equal(4, len(targets))
	assert.Equal(language.English, targets["en"])
	assert.Equal(language.French, targets["fr"])
	assert.Equal(language.German, targets["de"])
	assert.Equal(language.English, targets["en_au"].Parent().Parent())
	assert.ErrorIs(s.Errors(), &errs.ResolveError{})
	assert.ErrorContains(s.Errors(), "invalid")

	imports := s.ScanImports()

	assert.Equal(3, len(imports))
	assert.Equal("A", imports[0].Name)
	assert.Equal("B", imports[1].Name)
	assert.Equal("C", imports[2].Name)
}

func TestTypes(t *testing.T) {
	assert := assert.New(t)
	scope := pkg.NewScope()
	env := types.NewEnvironment()

	time := types.New("time", types.Int)

	env.Define("time", time)
	scope.Define("year", &types.Fn{
		In:  []types.Type{time},
		Out: types.Int,
	})
	scope.Define("itoa", &types.Fn{
		In:  []types.Type{types.Int},
		Out: types.String,
	})

	s := semantics("test/types.lcl", scope, env)

	s.ScanTypes()
	s.ScanFns()

	err := s.Errors().(*errs.SemanticError)

	assert.Equal(6, len(err.Reasons))
	assert.ErrorIs(err.Reasons[0], &errs.DuplicateError{})
	assert.ErrorIs(err.Reasons[1], &errs.ResolveError{})
	assert.ErrorIs(err.Reasons[2], &errs.DuplicateError{})
	assert.ErrorIs(err.Reasons[3], &errs.ResolveError{})
	assert.ErrorIs(err.Reasons[4], &errs.TypeError{})
	assert.ErrorIs(err.Reasons[5], &errs.TypeError{})

	exports := scope.Exports()

	fn := &types.Fn{}
	assert.IsType(fn, exports["Fn1"])
	assert.IsType(fn, exports["Fn2"])

	fn1 := exports["Fn1"].(*types.Fn)
	assert.ElementsMatch(fn1.In, []types.Type{time})
	assert.IsType(fn1.Out, types.Int)

	fn2 := exports["Fn2"].(*types.Fn)
	assert.ElementsMatch(fn2.In, []types.Type{time})
	assert.IsType(fn2.Out, types.String)
}

func TestImports(t *testing.T) {
	assert := assert.New(t)
	s := semantics("test/imports.lcl", nil, nil)

	s.ScanTags()
	s.ScanImports()
	s.ScanTypes()
	s.ScanFns()

	fmt.Println(s.Errors(), assert)
}

func TestSections(t *testing.T) {
	assert := assert.New(t)
	s := semantics("test/sections.lcl", nil, nil)

	s.ScanTags()
	s.ScanTypes()
	s.ScanFns()
	s.ScanSections()
	fmt.Println(s.Errors(), assert)
}
