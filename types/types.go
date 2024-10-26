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

type Extended struct {
	name string
	base Type
}

func (t *Extended) String() string {
	return t.name
}

func (t *Extended) IsRoot() bool {
	return false
}

func (t *Extended) Base() Type {
	return t.base
}

func (t *Extended) Assignable(o Type) bool {
	c, ok := o.(*Extended)
	if !ok {
		return false
	}

	return t.name == c.name
}

func (t *Extended) Comparable(o Type) bool {
	return t.base.Comparable(RootOf(o))
}

func (t *Extended) Convertible(o Type) bool {
	return t.base.Comparable(RootOf(o))
}

func (t *Extended) Operable(o Type, op Operation) bool {
	return t.base.Operable(o, op)
}

type indexer struct {
	Type
	Indexer
}

func New(name string, base Type) Type {
	var t Type = &Extended{
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
