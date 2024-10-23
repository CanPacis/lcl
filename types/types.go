package types

import (
	"fmt"
	"slices"
	"strings"
)

type Type interface {
	Name() string
	IsRoot() bool
	Base() Type
}

type typ struct {
	name string
	base Type
}

func (t *typ) Name() string {
	return t.name
}

func (t *typ) IsRoot() bool {
	return false
}

func (t *typ) Base() Type {
	return t.base
}

func New(name string, base Type) Type {
	return &typ{
		name: name,
		base: base,
	}
}

type constant struct {
	name string
}

func (t *constant) Name() string {
	return t.name
}

func (t *constant) IsRoot() bool {
	return true
}

func (t *constant) Base() Type {
	return nil
}

var String = &constant{"string"}
var Int = &constant{"int"}
var Float = &constant{"float"}
var Bool = &constant{"bool"}
var Empty = &constant{"empty"}

type List struct {
	Type Type
}

func (t *List) Name() string {
	return fmt.Sprintf("%s[]", t.Type.Name())
}

func (t *List) IsRoot() bool {
	return true
}

func (t *List) Base() Type {
	return nil
}

func NewList(t Type) *List {
	return &List{
		Type: t,
	}
}

type TypePair struct {
	Index int
	Name  string
	Type  Type
}

func NewPair(i int, name string, typ Type) TypePair {
	return TypePair{
		Index: i,
		Name:  name,
		Type:  typ,
	}
}

type Struct []TypePair

func (t *Struct) Name() string {
	fields := []string{}

	for _, pair := range *t {
		fields = append(fields, fmt.Sprintf("(%d %s %s)", pair.Index, pair.Name, pair.Type.Name()))
	}

	return fmt.Sprintf("struct {%s}", strings.Join(fields, " "))
}

func (t *Struct) IsRoot() bool {
	return true
}

func (t *Struct) Base() Type {
	return nil
}

func NewStruct(pairs ...TypePair) *Struct {
	s := Struct(pairs)
	slices.SortFunc(s, func(a, b TypePair) int {
		return a.Index - b.Index
	})

	return &s
}

type Fn struct {
	In  []Type
	Out Type
}

func (t *Fn) Name() string {
	in := []string{}

	for _, typ := range t.In {
		in = append(in, typ.Name())
	}

	return fmt.Sprintf("fn (%s) -> %s", strings.Join(in, " "), t.Out.Name())
}

func (t *Fn) IsRoot() bool {
	return true
}

func (t *Fn) Base() Type {
	return nil
}
