package pkg

import (
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/types"
)

type Scope struct {
	imports map[string]*Scope
	builtin map[string]types.Type
	objects map[string]types.Type

	importDefs map[string]*ast.IdentExpr
	fnDefs     map[string]*ast.FnDefStmt

	parent *Scope
}

func (s *Scope) Exports() map[string]types.Type {
	return s.objects
}

func (s *Scope) RegisterImport(def *ast.IdentExpr) error {
	if original, exists := s.importDefs[def.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Value,
		}
	}

	s.importDefs[def.Value] = def
	return nil
}

func (s *Scope) Import(name string, scope *Scope) {
	s.imports[name] = scope
}

func (s *Scope) RegisterFn(def *ast.FnDefStmt) error {
	if original, exists := s.fnDefs[def.Name.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Name.Value,
		}
	}

	s.fnDefs[def.Name.Value] = def
	return nil
}

func (s *Scope) Define(name string, typ types.Type) {
	s.objects[name] = typ
}

func (s Scope) Resolve(name string) (types.Type, error) {
	// If parent is not empty, check the local definitions first
	if s.parent != nil {
		typ, ok := s.objects[name]
		if ok {
			return typ, nil
		}

		return s.parent.Resolve(name)
	}
	// First check the builtins
	typ, ok := s.builtin[name]
	if ok {
		return typ, nil
	}

	typ, ok = s.objects[name]
	if ok {
		return typ, nil
	}

	return types.Invalid, &errs.ReferenceError{
		Err:   errs.ErrUnresolvedConstReference,
		Value: name,
	}
}

func NewScope() *Scope {
	return &Scope{
		imports:    make(map[string]*Scope),
		objects:    make(map[string]types.Type),
		importDefs: make(map[string]*ast.IdentExpr),
		fnDefs:     make(map[string]*ast.FnDefStmt),
		builtin: map[string]types.Type{
			"true":  types.Bool,
			"false": types.Bool,
		},
	}
}

func NewSubScope(parent *Scope) *Scope {
	return &Scope{
		imports:    make(map[string]*Scope),
		objects:    make(map[string]types.Type),
		builtin:    make(map[string]types.Type),
		importDefs: make(map[string]*ast.IdentExpr),
		fnDefs:     make(map[string]*ast.FnDefStmt),

		parent: parent,
	}
}
