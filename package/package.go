package pkg

import "github.com/CanPacis/lcl/types"

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
