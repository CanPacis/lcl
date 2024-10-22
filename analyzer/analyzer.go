package analyzer

import (
	"github.com/CanPacis/go-i18n/errs"
	pkg "github.com/CanPacis/go-i18n/package"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
	"golang.org/x/text/language"
)

type Analyzer struct {
	file    string
	ast     *ast.File
	checker *Checker

	errors []error

	Name    string
	Targets map[string]language.Tag
	TypeEnv *types.Environment
	Scope   *pkg.Scope
	Imports []*pkg.Package
}

func (a *Analyzer) error(err error) {
	a.errors = append(a.errors, err)
}

func (a *Analyzer) Run() error {
	a.scan()

	if len(a.errors) != 0 {
		return &errs.SemanticError{
			Reasons: a.errors,
			File:    a.file,
		}
	}

	return nil
}

func (a *Analyzer) scan() {
	a.Name = a.ast.For.In.Value

	for _, node := range a.ast.For.List {
		tag, err := language.Parse(node.Value)
		if err != nil {
			a.error(&errs.ResolveError{
				Kind:  errs.TARGET,
				Value: node.Value,
				Node:  node,
			})
		}

		a.Targets[node.Value] = tag
	}

	for _, node := range a.ast.Imports {
		for _, name := range node.List {
			pkg := a.resolveImport(name.Value)

			if pkg != nil {
				a.import_(pkg)
			} else {
				a.error(&errs.ResolveError{
					Kind:  errs.IMPORT,
					Value: name.Value,
					Node:  name,
				})
			}
		}
	}

	typeDefs := map[string]*ast.TypeDefStmt{}
	procDefs := map[string]*ast.ProcDefStmt{}

	for _, node := range a.ast.Stmts {
		switch node := node.(type) {
		case *ast.TypeDefStmt:
			if original, exists := typeDefs[node.Name.Value]; exists {
				a.error(&errs.DuplicateDefError{
					Name:     node.Name.Value,
					Original: original,
					Node:     node,
				})
			} else {
				typeDefs[node.Name.Value] = node
				a.TypeEnv.Define(node.Name.Value, nil)
			}
		case *ast.ProcDefStmt:
			if original, exists := procDefs[node.Name.Value]; exists {
				a.error(&errs.DuplicateDefError{
					Name:     node.Name.Value,
					Original: original,
					Node:     node,
				})
			} else {
				procDefs[node.Name.Value] = node
				a.Scope.Define(node.Name.Value, nil)
			}
		}
	}

	for name, stmt := range typeDefs {
		typ, err := a.checker.ResolveType(stmt.Type)
		if err != nil {
			a.error(err)
		}
		a.TypeEnv.Define(name, typ)
	}

	for name, stmt := range procDefs {
		out, err := a.checker.ResolveExpr(stmt.Body)
		if err != nil {
			a.error(err)
		}
		a.Scope.Define(name, &types.Proc{
			In:  types.Empty,
			Out: out,
		})
	}
}

func (a *Analyzer) resolveImport(name string) *pkg.Package {
	switch name {
	case "List":
		return pkg.ListPkg
	default:
		return nil
	}
}

func (a *Analyzer) import_(pkg *pkg.Package) {
	a.Imports = append(a.Imports, pkg)
	a.TypeEnv.Import(pkg.Name, pkg.TypEnv)
	a.Scope.Import(pkg.Name, pkg.Scope)
}

func New(file *parser.File, ast *ast.File) *Analyzer {
	a := &Analyzer{
		file: file.Name,
		ast:  ast,

		Targets: make(map[string]language.Tag),
		Scope:   pkg.NewScope(),
		TypeEnv: types.NewEnvironment(),
	}
	a.checker = NewChecker(a.Scope, a.TypeEnv)

	return a
}
