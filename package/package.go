package pkg

import "github.com/CanPacis/go-i18n/types"

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
