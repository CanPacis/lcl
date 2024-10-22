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
		tag, err := s.checker.RegisterTarget(node)
		if err != nil {
			s.error(err)
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
	defs := map[string]*ast.FnDefStmt{}

	for _, node := range s.ast.Stmts {
		switch node := node.(type) {
		case *ast.FnDefStmt:
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

func (s *Semantics) extractKeyEntry(entry *ast.KeyEntry) *Key {
	key := &Key{
		Name:   entry.Name.Value,
		Fields: make(map[language.Tag]string),
	}

	for _, field := range entry.Fields {
		tag, err := s.checker.LookupTag(field.Tag)
		if err != nil {
			s.error(err)
		}
		key.Fields[tag] = field.Value.Value
	}
	for name, tag := range s.checker.tags {
		_, ok := key.Fields[tag]
		if !ok {
			s.error(&errs.TargetMismatchError{
				Target:  name,
				Tag:     tag,
				Missing: true,
				Node:    entry,
			})
		}
	}

	return key
}

func (s *Semantics) extractTemplateEntry(entry *ast.TemplateEntry) *Template {
	typ, err := s.checker.ResolveType(entry.Type)
	if err != nil {
		s.error(err)
	}

	template := &Template{
		Name:   entry.Name.Value,
		Type:   typ,
		Fields: make(map[string]int),
	}

	return template
}

func (s *Semantics) extractSection(stmt *ast.SectionStmt) *Section {
	section := &Section{
		Name:      stmt.Name.Value,
		Keys:      make(map[string]*Key),
		Templates: make(map[string]*Template),
	}

	for _, entry := range stmt.Body {
		switch entry := entry.(type) {
		case *ast.KeyEntry:
			key := s.extractKeyEntry(entry)
			section.Keys[entry.Name.Value] = key
		case *ast.TemplateEntry:
			template := s.extractTemplateEntry(entry)
			section.Templates[entry.Name.Value] = template
		case *ast.SectionStmt:
			section.Sections = append(section.Sections, s.extractSection(entry))
		}
	}

	return section
}

func (s *Semantics) ScanSections() []*Section {
	sections := []*Section{}

	for _, node := range s.ast.Stmts {
		switch node := node.(type) {
		case *ast.SectionStmt:
			sections = append(sections, s.extractSection(node))
		}
	}

	return sections
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
