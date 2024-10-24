package types

import (
	"fmt"
	"slices"
	"strings"
)

type Type interface {
	String() string
	IsRoot() bool
	Base() Type
}

type typ struct {
	name string
	base Type
}

func (t *typ) String() string {
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

func (t *constant) String() string {
	return t.name
}

func (t *constant) IsRoot() bool {
	return true
}

func (t *constant) Base() Type {
	return nil
}

var (
	Invalid = &constant{"invalid"}

	Bool   = &constant{"bool"}
	I8     = &constant{"i8"}
	I16    = &constant{"i16"}
	I32    = &constant{"i32"}
	I64    = &constant{"i64"}
	U8     = &constant{"u8"}
	U16    = &constant{"u16"}
	U32    = &constant{"u32"}
	U64    = &constant{"u64"}
	F32    = &constant{"f32"}
	F64    = &constant{"f64"}
	Byte   = New("byte", U8)
	Rune   = New("rune", U32)
	String = New("string", NewList(Rune))
)

type List struct {
	Type Type
}

func (t *List) String() string {
	return fmt.Sprintf("%s[]", t.Type.String())
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

func (t *Struct) String() string {
	fields := []string{}

	for _, pair := range *t {
		fields = append(fields, fmt.Sprintf("(%d %s %s)", pair.Index, pair.Name, pair.Type.String()))
	}

	return fmt.Sprintf("{%s}", strings.Join(fields, " "))
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

type Template struct {
	In []Type
}

func (t *Template) Name() string {
	in := []string{}

	for _, typ := range t.In {
		in = append(in, typ.String())
	}

	return fmt.Sprintf("template (%s)", strings.Join(in, " "))
}

func (t *Template) IsRoot() bool {
	return true
}

func (t *Template) Base() Type {
	return nil
}

func NewTemplate(in []Type) *Template {
	return &Template{
		In: in,
	}
}

type Fn struct {
	In  []Type
	Out Type
}

func (t *Fn) String() string {
	in := []string{}

	for _, typ := range t.In {
		in = append(in, typ.String())
	}

	return fmt.Sprintf("fn (%s) -> %s", strings.Join(in, " "), t.Out.String())
}

func (t *Fn) IsRoot() bool {
	return true
}

func (t *Fn) Base() Type {
	return nil
}
