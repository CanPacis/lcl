package analyzer_test

import (
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

type AnalyzerCase struct {
	In       string
	Test     func(*analyzer.Semantics, *assert.Assertions)
	Contains string
}

func (c *AnalyzerCase) Run(assert *assert.Assertions) {
	file := file(c.In)
	ast, err := parser.New(file).Parse()
	if err != nil {
		panic(err)
	}
	sem := analyzer.New(file, ast)
	if c.Test != nil {
		c.Test(sem, assert)
	}
	if len(c.Contains) > 0 {
		assert.ErrorContains(sem.Errors(), c.Contains)
	}
}

func TestAnalyzer(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			In: `for (en fr de) in i18n`,
			Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
				assert.Equal("i18n", s.ScanName())
				targets := s.ScanTargets()
				assert.Equal(3, len(targets))
				assert.Equal(language.English, targets[0])
				assert.Equal(language.French, targets[1])
				assert.Equal(language.German, targets[2])
			},
		},
		&AnalyzerCase{
			In: `for (invalid) in i18n`,
			Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
				targets := s.ScanTargets()
				assert.Equal(1, len(targets))
				assert.Equal(language.Und, targets[0])
			},
			Contains: errs.Unresolved,
		},
		&AnalyzerCase{
			In: `for (en) in i18n import (A B C)`,
			Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
				imports := s.ScanImports()
				assert.Equal(3, len(imports))
				assert.Equal("A", imports[0].Name)
				assert.Equal("B", imports[1].Name)
				assert.Equal("C", imports[2].Name)
			},
		},
	}

	Run(tests, t)
}

func TestTypeAnalysis(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			In: `for (en) in i18n
				type Test bool
				type Test int
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
			},
			Contains: errs.Duplicate,
		},
		&AnalyzerCase{
			In: `for (en) in i18n
				type Test undefined
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
			},
			Contains: errs.Unresolved,
		},
		&AnalyzerCase{
			In: `for (en) in i18n
				proc test .
				proc test .
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
				s.ScanProcs()
			},
			Contains: errs.Duplicate,
		},
		&AnalyzerCase{
			In: `for (en) in i18n
				proc test undefined
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
				s.ScanProcs()
			},
			Contains: errs.Unresolved,
		},
		&AnalyzerCase{
			In: `for (en) in i18n
				proc p1 itoa(year(.))
				proc p2 year(itoa(.))
				proc p3 year(.)
			`,
			Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
				s.ScanTypes()
				scope := s.ScanProcs()

				exports := scope.Exports()
				p1 := exports["p1"].(*types.Proc)
				p2 := exports["p2"].(*types.Proc)
				p3 := exports["p3"].(*types.Proc)

				assert.Equal(types.Time, p1.In)
				assert.Equal(types.String, p1.Out)

				assert.Equal(types.Int, p2.In)
				assert.Equal(types.Empty, p2.Out)

				assert.Equal(types.Time, p3.In)
				assert.Equal(types.Int, p3.Out)
			},
			Contains: errs.NotAssignable,
		},
		&AnalyzerCase{
			In: `for (en) in i18n
				proc test year((. == 4))
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
				s.ScanProcs()
			},
			Contains: errs.NotInferrable,
		},
	}

	Run(tests, t)
}

func TestForeignTypeAnalysis(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			In: `for (en) in i18n
				import A

				proc p A.B(.)
			`,
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanImports()
				s.ScanTypes()
				s.ScanProcs()

				fmt.Println(s.Errors())
			},
		},
	}

	Run(tests, t)
}
