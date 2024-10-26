package types

type Constant struct {
	name  string
	canop bool
}

func (t *Constant) String() string {
	return t.name
}

func (t *Constant) IsRoot() bool {
	return true
}

func (t *Constant) Base() Type {
	return nil
}

func (t *Constant) Assignable(o Type) bool {
	c, ok := o.(*Constant)
	if !ok {
		return false
	}

	return t.name == c.name
}

func (t *Constant) Comparable(o Type) bool {
	c, ok := RootOf(o).(*Constant)
	if !ok {
		return false
	}
	return t.name == c.name
}

func (t *Constant) Convertible(o Type) bool {
	return t.Comparable(RootOf(o))
}

func (t *Constant) Operable(o Type, op Operation) bool {
	if !t.canop {
		return false
	}

	// TODO
	return t.Comparable(o)
}

var (
	Invalid = &Constant{"invalid", false}

	Bool   = &Constant{"bool", false}
	I8     = &Constant{"i8", true}
	I16    = &Constant{"i16", true}
	I32    = &Constant{"i32", true}
	I64    = &Constant{"i64", true}
	U8     = &Constant{"u8", true}
	U16    = &Constant{"u16", true}
	U32    = &Constant{"u32", true}
	U64    = &Constant{"u64", true}
	F32    = &Constant{"f32", true}
	F64    = &Constant{"f64", true}
	Int    = New("int", I32)
	Uint   = New("uint", U8)
	Byte   = New("byte", U8)
	Rune   = New("rune", U32)
	String = New("string", NewList(Rune))
)
