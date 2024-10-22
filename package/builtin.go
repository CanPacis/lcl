package pkg

import "github.com/CanPacis/go-i18n/types"

var ListPkg = &Package{
	Name:   "List",
	TypEnv: types.NewEnvironment(),
	Scope:  NewScope(),
}

func init() {
	ListPkg.Scope.Define("Join", &types.Proc{
		In:  &types.List{Type: types.String},
		Out: types.String,
	})
}
