package parser

import (
	"io"
)

type Context int

const (
	TOP_LEVEL Context = iota
	STATEMENT
	ENTRY
	EXPRESSION
	TYPE_EXPRESSION
	SEQUENCE
)

var ctxMap = map[Context]string{
	TOP_LEVEL:  "top level",
	STATEMENT:  "statement",
	ENTRY:      "entry",
	EXPRESSION: "expression",
	SEQUENCE:   "sequence",
}

func (c Context) String() string {
	return ctxMap[c]
}

type File struct {
	Name   string
	source io.Reader
}

func NewFile(name string, r io.Reader) *File {
	return &File{
		Name:   name,
		source: r,
	}
}
