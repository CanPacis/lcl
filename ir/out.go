package ir

import (
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/types"
)

type Definition struct {
	Name     string
	Exported bool
}

func NewDefinition(name string, exported bool) *Definition {
	return &Definition{
		Name:     name,
		Exported: exported,
	}
}

type FnDef struct {
	*Definition
	Type *types.Fn
	Stmt *ast.FnDefStmt
}

type TypeDef struct {
	*Definition
	Type      types.Type
	IsSection bool
}
