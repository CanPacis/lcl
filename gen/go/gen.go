package gogen

import (
	pkg "github.com/CanPacis/lcl/package"
	"github.com/CanPacis/lcl/types"
)

type Generator struct {
	config *Config
	scope  *pkg.Scope
	env    *types.Environment
}

func New(scope *pkg.Scope, env *types.Environment, options ...func(*Config)) *Generator {
	config := &Config{
		root:  "root",
		local: "Local",
		fn:    "fn",
	}

	for _, option := range options {
		option(config)
	}

	return &Generator{
		config: config,
		scope:  scope,
		env:    env,
	}
}
