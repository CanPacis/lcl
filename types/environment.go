package types

import (
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser/ast"
)

type Environment struct {
	imports  map[string]*Environment
	builtins map[string]Type
	types    map[string]Type

	importDefs map[string]*ast.IdentExpr
	typeDefs   map[string]*ast.TypeDefStmt
}

func (e *Environment) Exports() map[string]Type {
	return e.types
}

func (e *Environment) RegisterImport(def *ast.IdentExpr) error {
	if original, exists := e.importDefs[def.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Value,
		}
	}

	e.importDefs[def.Value] = def
	return nil
}

func (e *Environment) Import(name string, env *Environment) {
	e.imports[name] = env
}

func (e Environment) lookupType(name string) (Type, bool) {
	typ, ok := e.builtins[name]
	if ok {
		return typ, ok
	}

	typ, ok = e.types[name]
	if ok {
		return typ, ok
	}
	return Invalid, false
}

func (e *Environment) IsTypeRegistered(expr ast.TypeExpr) bool {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		_, ok := e.typeDefs[expr.Value]
		return ok
	default:
		return false
	}
}

func (e *Environment) RegisterType(def *ast.TypeDefStmt) error {
	if original, exists := e.typeDefs[def.Name.Value]; exists {
		return &errs.ReferenceError{
			Err:      errs.ErrDuplicateDefinition,
			Node:     def,
			Original: original,
			Value:    def.Name.Value,
		}
	}

	if typ, exists := e.builtins[def.Name.Value]; exists {
		return &errs.TypeError{
			Err:  errs.ErrBuiltinOverride,
			Node: def,
			Type: typ,
		}
	}

	e.typeDefs[def.Name.Value] = def
	return nil
}

func (e *Environment) DefineType(name string, typ Type) {
	e.types[name] = typ
}

func (e Environment) ResolveType(expr ast.TypeExpr) (Type, error) {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		typ, ok := e.lookupType(expr.Value)
		if ok {
			return typ, nil
		}

		return Invalid, &errs.ReferenceError{
			Err:   errs.ErrUnresolvedTypeReference,
			Node:  expr,
			Value: expr.Value,
		}
	case *ast.ImportExpr:
		foreign, ok := e.imports[expr.Left.Value]
		if !ok {
			return Invalid, &errs.ReferenceError{
				Err:   errs.ErrUnresolvedImportReference,
				Value: expr.Left.Value,
			}
		}

		return foreign.ResolveType(expr.Right)
	case *ast.StructLitExpr:
		pairs := []TypePair{}
		var err error

		for _, field := range expr.Fields {
			var typ Type
			typ, err = e.ResolveType(field.Type)
			if err != nil {
				pairs = append(pairs, NewPair(field.Index, field.Name.Value, Invalid))
			} else {
				pairs = append(pairs, NewPair(field.Index, field.Name.Value, typ))
			}

		}

		return NewStruct(pairs...), err
	case *ast.ListTypeExpr:
		typ, err := e.ResolveType(expr.Type)
		if err != nil {
			return NewList(Invalid), err
		}

		return NewList(typ), err
	default:
		return Invalid, &errs.TypeError{
			Err: errs.ErrInvalidType,
		}
	}
}

func NewEnvironment() *Environment {
	return &Environment{
		imports:    make(map[string]*Environment),
		types:      make(map[string]Type),
		typeDefs:   make(map[string]*ast.TypeDefStmt),
		importDefs: make(map[string]*ast.IdentExpr),
		builtins: map[string]Type{
			"bool":   Bool,
			"i8":     I8,
			"i16":    I16,
			"i32":    I32,
			"i64":    I64,
			"u8":     U8,
			"u16":    U16,
			"u32":    U32,
			"u64":    U64,
			"f32":    F32,
			"f64":    F64,
			"byte":   Byte,
			"rune":   Rune,
			"string": String,
		},
	}
}
