package types

import (
	"fmt"
	"slices"
	"strings"
)

type Type interface {
	Name() string
}

type constant struct {
	name string
}

func (t *constant) Name() string {
	return t.name
}

var String = &constant{"string"}
var Int = &constant{"int"}
var Float = &constant{"float"}
var Bool = &constant{"bool"}
var Time = &constant{"time"}
var Self = &constant{"self"}
var Empty = &constant{"empty"}

type List struct {
	Type Type
}

func (t *List) Name() string {
	return fmt.Sprintf("%s[]", t.Type.Name())
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

func NewStruct(pairs ...TypePair) *Struct {
	s := Struct(pairs)
	slices.SortFunc(s, func(a, b TypePair) int {
		return a.Index - b.Index
	})

	return &s
}

type Fn struct {
	In  Type
	Out Type
}

func (t *Fn) Name() string {
	return fmt.Sprintf("fn (%s -> %s)", t.In.Name(), t.Out.Name())
}
