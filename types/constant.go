package types

type constant struct {
	name  string
	canop bool
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

func (t *constant) Assignable(o Type) bool {
	c, ok := o.(*constant)
	if !ok {
		return false
	}

	return t.name == c.name
}

func (t *constant) Comparable(o Type) bool {
	c, ok := RootOf(o).(*constant)
	if !ok {
		return false
	}
	return t.name == c.name
}

func (t *constant) Convertible(o Type) bool {
	return t.Comparable(RootOf(o))
}

func (t *constant) Operable(o Type, op Operation) bool {
	if !t.canop {
		return false
	}

	// TODO
	return t.Comparable(o)
}

var (
	Invalid = &constant{"invalid", false}

	Bool   = &constant{"bool", false}
	I8     = &constant{"i8", true}
	I16    = &constant{"i16", true}
	I32    = &constant{"i32", true}
	I64    = &constant{"i64", true}
	U8     = &constant{"u8", true}
	U16    = &constant{"u16", true}
	U32    = &constant{"u32", true}
	U64    = &constant{"u64", true}
	F32    = &constant{"f32", true}
	F64    = &constant{"f64", true}
	Int    = New("int", I32)
	Uint   = New("uint", U8)
	Byte   = New("byte", U8)
	Rune   = New("rune", U32)
	String = New("string", NewList(Rune))
)
