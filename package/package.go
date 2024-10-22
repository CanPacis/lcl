package pkg

import (
	"github.com/CanPacis/go-i18n/types"
)

type Scope struct {
	imports map[string]*Scope
	builtin map[string]types.Type

	objects map[string]types.Type
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

func (s Scope) Lookup(name, prefix string) (types.Type, bool) {
	if len(prefix) > 0 {
		scope := s.imports[prefix]
		if scope == nil {
			return nil, false
		}

		return scope.Lookup(name, "")
	}

	typ, ok := s.builtin[name]
	if ok {
		return typ, ok
	}

	typ, ok = s.objects[name]
	if ok {
		return typ, ok
	}

	return typ, ok
}

func NewScope() *Scope {
	return &Scope{
		imports: make(map[string]*Scope),
		objects: make(map[string]types.Type),
		builtin: map[string]types.Type{
			"true":  types.Bool,
			"false": types.Bool,
			"itoa": &types.Proc{
				In:  types.Int,
				Out: types.String,
			},
			"year": &types.Proc{
				In:  types.Time,
				Out: types.Int,
			},
		},
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
