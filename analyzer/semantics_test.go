package analyzer_test

import (
	"testing"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

type AnalyzerCase struct {
	File     *parser.File
	Test     func(*analyzer.Semantics, *assert.Assertions)
	Contains string
}

func (c *AnalyzerCase) Run(assert *assert.Assertions) {
	ast, err := parser.New(c.File).Parse()
	if err != nil {
		panic(FormatError(err))
	}
	sem := analyzer.New(c.File, ast)
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
			File: Test["TestAnalyzer0"],
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
			File: Test["TestAnalyzer1"],
			Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
				targets := s.ScanTargets()
				assert.Equal(1, len(targets))
				assert.Equal(language.Und, targets[0])
			},
			Contains: errs.Unresolved,
		},
		&AnalyzerCase{
			File: Test["TestAnalyzer2"],
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

func TestLocalTypeAnalysis(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			File: Test["TestLocalTypeAnalysis0"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
			},
			Contains: errs.Duplicate,
		},
		&AnalyzerCase{
			File: Test["TestLocalTypeAnalysis1"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
			},
			Contains: errs.Unresolved,
		},
		&AnalyzerCase{
			File: Test["TestLocalTypeAnalysis2"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
				s.ScanFns()
			},
			Contains: errs.Duplicate,
		},
		&AnalyzerCase{
			File: Test["TestLocalTypeAnalysis3"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTypes()
				s.ScanFns()
			},
			Contains: errs.Unresolved,
		},
		// &AnalyzerCase{
		// 	File: Test["TestLocalTypeAnalysis4"],
		// 	Test: func(s *analyzer.Semantics, assert *assert.Assertions) {
		// 		s.ScanTypes()
		// 		scope := s.ScanFns()

		// 		exports := scope.Exports()
		// 		p1 := exports["p1"].(*types.Fn)
		// 		p2 := exports["p2"].(*types.Fn)
		// 		p3 := exports["p3"].(*types.Fn)

		// 		assert.Equal([]types.Type{types.Time}, p1.In)
		// 		assert.Equal(types.String, p1.Out)

		// 		assert.Equal([]types.Type{types.Time}, p2.In)
		// 		assert.Equal(types.Int, p2.Out)

		// 		assert.Equal([]types.Type{types.Time}, p3.In)
		// 		assert.Equal(types.Int, p3.Out)
		// 	},
		// 	Contains: errs.NotAssignable,
		// },
	}

	Run(tests, t)
}

func TestForeignTypeAnalysis(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			File: Test["TestForeignTypeAnalysis0"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanImports()
				s.ScanTypes()
				s.ScanFns()
			},
		},
	}

	Run(tests, t)
}

func TestSections(t *testing.T) {
	tests := []Runner{
		&AnalyzerCase{
			File: Test["TestSections0"],
			Test: func(s *analyzer.Semantics, a *assert.Assertions) {
				s.ScanTargets()
				s.ScanTypes()
				s.ScanFns()
				s.ScanSections()
			},
		},
	}

	Run(tests, t)
}
