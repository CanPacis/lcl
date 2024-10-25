package types

type Type interface {
	String() string
	IsRoot() bool
	Base() Type
	Assignable(Type) bool
	Comparable(Type) bool
	Convertible(Type) bool
	Operable(Type, Operation) bool
}

type Indexer interface {
	Index(any) (Type, bool)
}

type Operation int

const (
	Noop Operation = iota
	Addition
	Subtraction
	Division
	Multiplication
	Exponent
)

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

func (t *typ) Assignable(o Type) bool {
	c, ok := o.(*typ)
	if !ok {
		return false
	}

	return t.name == c.name
}

func (t *typ) Comparable(o Type) bool {
	return t.base.Comparable(RootOf(o))
}

func (t *typ) Convertible(o Type) bool {
	return t.base.Comparable(RootOf(o))
}

func (t *typ) Operable(o Type, op Operation) bool {
	return false
}

type indexer struct {
	Type
	Indexer
}

func New(name string, base Type) Type {
	var t Type = &typ{
		name: name,
		base: base,
	}

	i, ok := base.(Indexer)
	if ok {
		t = &indexer{
			Type:    t,
			Indexer: i,
		}
	}

	return t
}

func RootOf(t Type) Type {
	if t.IsRoot() {
		return t
	}

	return RootOf(t.Base())
}
