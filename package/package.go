package pkg

import (
	"github.com/CanPacis/go-i18n/types"
)

type Scope struct {
	imports map[string]*Scope
	builtin map[string]types.Type
	objects map[string]types.Type

	parent *Scope
}

func (s *Scope) Exports() map[string]types.Type {
	return s.objects
}

func (s *Scope) Import(name string, scope *Scope) {
	s.imports[name] = scope
}

func (s *Scope) Define(name string, typ types.Type) {
	s.objects[name] = typ
}

func (s Scope) Lookup(name string) (types.Type, bool) {
	if s.parent != nil {
		typ, ok := s.objects[name]
		if ok {
			return typ, ok
		}

		return s.parent.Lookup(name)
	}

	typ, ok := s.builtin[name]
	if ok {
		return typ, ok
	}

	typ, ok = s.objects[name]
	if ok {
		return typ, ok
	}
	return types.Empty, false
}

func NewScope() *Scope {
	return &Scope{
		imports: make(map[string]*Scope),
		objects: make(map[string]types.Type),
		builtin: map[string]types.Type{
			"true":  types.Bool,
			"false": types.Bool,
			"itoa": &types.Fn{
				In:  []types.Type{types.Int},
				Out: types.String,
			},
		},
	}
}

func NewSubScope(parent *Scope) *Scope {
	return &Scope{
		imports: make(map[string]*Scope),
		objects: make(map[string]types.Type),
		builtin: make(map[string]types.Type),

		parent: parent,
	}
}

type Package struct {
	Name   string
	TypEnv *types.Environment
	Scope  *Scope
}

func New(name string) *Package {
	return &Package{
		Name:   name,
		TypEnv: types.NewEnvironment(),
		Scope:  NewScope(),
	}
}
