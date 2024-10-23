package types

type Environment struct {
	imports map[string]*Environment
	builtin map[string]Type
	types   map[string]Type
}

func (e *Environment) Exports() map[string]Type {
	return e.types
}

func (e *Environment) Import(name string, env *Environment) {
	e.imports[name] = env
}

func (e *Environment) Define(name string, typ Type) {
	e.types[name] = typ
}

func (e Environment) Lookup(name string) (Type, bool) {
	typ, ok := e.builtin[name]
	if ok {
		return typ, ok
	}

	typ, ok = e.types[name]
	if ok {
		return typ, ok
	}
	return Empty, ok
}

func NewEnvironment() *Environment {
	return &Environment{
		imports: map[string]*Environment{},
		types:   map[string]Type{},
		builtin: map[string]Type{
			"string": String,
			"int":    Int,
			"float":  Float,
			"bool":   Bool,
		},
	}
}
