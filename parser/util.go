package parser

import (
	"io"
)

type ParseContext int

const (
	TOP_LEVEL ParseContext = iota
	STATEMENT
	ENTRY
	EXPRESSION
	TYPE_EXPRESSION
	SEQUENCE
)

var ctxMap = map[ParseContext]string{
	TOP_LEVEL:  "top level",
	STATEMENT:  "statement",
	ENTRY:      "entry",
	EXPRESSION: "expression",
	SEQUENCE:   "sequence",
}

func (c ParseContext) String() string {
	return ctxMap[c]
}

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) Push(v T) {
	s.values = append(s.values, v)
}

func (s *Stack[T]) Pop() T {
	if len(s.values) == 0 {
		var t T
		return t
	}

	popped := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return popped
}

func (s Stack[T]) Last() T {
	if len(s.values) == 0 {
		var t T
		return t
	}

	return s.values[len(s.values)-1]
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
