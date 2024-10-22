package pkg

import "github.com/CanPacis/go-i18n/types"

var ListPkg = &Package{
	Name:   "List",
	TypEnv: types.NewEnvironment(),
	Scope:  NewScope(),
}

func init() {
	ListPkg.Scope.Define("Join", &types.Fn{
		In: []types.Type{
			&types.List{Type: types.String},
		},
		Out: types.String,
	})
}
