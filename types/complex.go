package types

import (
	"fmt"
	"slices"
	"strings"
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

func (t *List) Index(v any) (Type, bool) {
	_, ok := v.(int)
	if !ok {
		return Invalid, false
	}
	return t.Type, true
}

func (t *List) Assignable(o Type) bool {
	c, ok := o.(*List)
	if !ok {
		return false
	}

	return t.Type.Assignable(c.Type)
}

func (t *List) Comparable(o Type) bool {
	c, ok := RootOf(o).(*List)
	if !ok {
		return false
	}

	return t.Type.Comparable(c.Type)
}

func (t *List) Convertible(o Type) bool {
	return t.Comparable(RootOf(o))
}

func (t *List) Operable(o Type, op Operation) bool {
	switch op {
	case Addition:
		// TODO: not exactly right
		return t.Comparable(o)
	default:
		return false
	}
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

func (t *Struct) Index(v any) (Type, bool) {
	member, ok := v.(string)
	if !ok {
		return Invalid, false
	}

	for _, pair := range *t {
		if pair.Name == member {
			return pair.Type, true
		}
	}
	return Invalid, false
}

func (t *Struct) Assignable(o Type) bool {
	// TODO
	return false
}

func (t *Struct) Comparable(o Type) bool {
	_, ok := RootOf(o).(*Struct)
	if !ok {
		return false
	}
	// TODO
	return false
}

func (t *Struct) Convertible(o Type) bool {
	// TODO
	return t.Comparable(RootOf(o))
}

func (t *Struct) Operable(o Type, op Operation) bool {
	return false
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

func (t *Template) String() string {
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

func (t *Template) Assignable(o Type) bool {
	// TODO
	return false
}

func (t *Template) Comparable(o Type) bool {
	_, ok := RootOf(o).(*Template)
	return ok
}

func (t *Template) Convertible(o Type) bool {
	return false
}

func (t *Template) Operable(o Type, op Operation) bool {
	// TODO: maybe concat?
	return false
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

func (t *Fn) Assignable(o Type) bool {
	// TODO
	return false
}

func (t *Fn) Comparable(o Type) bool {
	_, ok := RootOf(o).(*Fn)
	return ok
}

func (t *Fn) Convertible(o Type) bool {
	return false
}

func (t *Fn) Operable(o Type, op Operation) bool {
	return false
}
