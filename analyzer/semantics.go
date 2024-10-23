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
	return s.ast.Decl.Name.Value
}

func (s *Semantics) ScanTags() map[string]language.Tag {
	for _, node := range s.ast.Decl.Targets {
		if err := s.checker.RegisterTarget(node); err != nil {
			s.error(err)
		}
	}
	return s.checker.tags
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

func (s *Semantics) ScanFns() *pkg.Scope {
	defs := map[string]*ast.FnDefStmt{}

	for _, node := range s.ast.Stmts {
		switch node := node.(type) {
		case *ast.FnDefStmt:
			if err := s.checker.RegisterFn(node); err != nil {
				s.error(err)
			} else {
				defs[node.Name.Value] = node
			}
		}
	}

	for name, def := range defs {
		in := []types.Type{}
		s.checker.BeginCtx(FN_BODY)
		s.checker.PushScope()

		for _, param := range def.Params {
			typ, err := s.checker.ResolveType(param.Type)
			if err != nil {
				s.error(err)
			}
			in = append(in, typ)
			s.checker.Scope().Define(param.Name.Value, typ)
		}

		typ, err := s.checker.ResolveExpr(def.Body)
		if err != nil {
			s.error(err)
		}
		s.checker.PopScope()
		s.checker.EndCtx()

		s.checker.Scope().Define(name, &types.Fn{
			In:  in,
			Out: typ,
		})
	}
	return s.checker.Scope()
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
		// TODO: validate the expression
		// key.Fields[tag] = field.Value
		key.Fields[tag] = ""
	}

	for name, tag := range s.checker.tags {
		_, ok := key.Fields[tag]
		if !ok {
			s.error(&errs.TargetError{
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
	params := []types.Type{}
	for _, param := range entry.Params {
		typ, err := s.checker.ResolveType(param.Name)
		if err != nil {
			s.error(err)
		}
		params = append(params, typ)
	}

	template := &Template{
		Name:   entry.Name.Value,
		Params: params,
		Fields: make(map[language.Tag]int),
	}

	for _, field := range entry.Fields {
		tag, err := s.checker.LookupTag(field.Tag)
		if err != nil {
			s.error(err)
		}

		// TODO: validate the expression, and figure out what to do with field value
		template.Fields[tag] = 0
	}

	for name, tag := range s.checker.tags {
		_, ok := template.Fields[tag]
		if !ok {
			s.error(&errs.TargetError{
				Target:  name,
				Tag:     tag,
				Missing: true,
				Node:    entry,
			})
		}
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

func New(file *parser.File, ast *ast.File, checker *Checker) *Semantics {
	return &Semantics{
		file:    file.Name,
		ast:     ast,
		checker: checker,
	}
}
