package types

import "fmt"

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

// type String struct {
// }

// func (t *String) Name() string {
// 	return "string"
// }

// type Int struct {
// }

// func (t *Int) Name() string {
// 	return "int"
// }

// type Float struct {
// }

// func (t *Float) Name() string {
// 	return "float"
// }

// type Bool struct {
// }

// func (t *Bool) Name() string {
// 	return "bool"
// }

type List struct {
	Type Type
}

func (t *List) Name() string {
	return fmt.Sprintf("%s[]", t.Type.Name())
}

// type Time struct {
// }

// func (t *Time) Name() string {
// 	return "time"
// }

type Struct struct {
	Fields map[string]Type
}

func (t *Struct) Name() string {
	return "struct"
}

func NewStruct(fields map[string]Type) *Struct {
	return &Struct{
		Fields: fields,
	}
}

type Proc struct {
	In  Type
	Out Type
}

func (t *Proc) Name() string {
	return fmt.Sprintf("proc %s -> %s", t.In.Name(), t.Out.Name())
}
