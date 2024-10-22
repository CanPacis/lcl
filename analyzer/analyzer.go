package analyzer

import (
	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
	"golang.org/x/text/language"
)

type Semantics struct {
	file    string
	ast     *ast.File
	checker *Checker

	errors []error
}

func (s *Semantics) error(err error) {
	if err != nil {
		s.errors = append(s.errors, err)
	}
}

func (s Semantics) Errors() error {
	if len(s.errors) == 0 {
		return nil
	}
	return errs.NewSemanticError(s.errors, s.file)
}

func (s Semantics) ScanName() string {
	return s.ast.For.In.Value
}

func (s *Semantics) ScanTargets() []language.Tag {
	targets := []language.Tag{}

	for _, node := range s.ast.For.List {
		tag, err := language.Parse(node.Value)
		if err != nil {
			s.error(&errs.ResolveError{
				Value: node.Value,
				Kind:  errs.TARGET,
				Node:  node,
			})
		}
		targets = append(targets, tag)
	}
	return targets
}

func (s *Semantics) ScanImports() []*pkg.Package {
	imports := []*pkg.Package{}

	for _, node := range s.ast.Imports {
		for _, ident := range node.List {
			// TODO: resolve the import
			imports = append(imports, pkg.New(ident.Value))
		}
	}

	return imports
}

func (s *Semantics) ScanTypes() *types.Environment {
	defs := map[string]*ast.TypeDefStmt{}

	for _, node := range s.ast.Stmts {
		switch node := node.(type) {
		case *ast.TypeDefStmt:
			if err := s.checker.RegisterType(node); err != nil {
				s.error(err)
			} else {
				defs[node.Name.Value] = node
			}
		}
	}

	for name, def := range defs {
		typ, err := s.checker.ResolveType(def.Type)
		if err != nil {
			s.error(err)
		}
		s.checker.env.Define(name, typ)
	}

	return s.checker.env
}

func (s *Semantics) ScanProcs() *pkg.Scope {
	defs := map[string]*ast.ProcDefStmt{}

	for _, node := range s.ast.Stmts {
		switch node := node.(type) {
		case *ast.ProcDefStmt:
			if err := s.checker.RegisterProc(node); err != nil {
				s.error(err)
			} else {
				defs[node.Name.Value] = node
			}
		}
	}

	for name, def := range defs {
		s.checker.Begin(PROC_BODY)
		typ, err := s.checker.ResolveExpr(def.Body)
		if err != nil {
			s.error(err)
		}
		s.checker.End()

		s.checker.scope.Define(name, &types.Proc{
			In:  s.checker.self,
			Out: typ,
		})
		s.checker.self = types.Empty
	}
	return s.checker.scope
}

func (s *Semantics) Scan() error {
	return s.Errors()
}

func New(file *parser.File, ast *ast.File) *Semantics {
	return &Semantics{
		file:    file.Name,
		ast:     ast,
		checker: NewChecker(pkg.NewScope(), types.NewEnvironment()),
	}
}
