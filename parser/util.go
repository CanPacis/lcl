package parser

import "io"

type ParseContext int

const (
	TOP_LEVEL ParseContext = iota
	STATEMENT
	ENTRY
	EXPRESSION
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

type ContextFrame struct {
	frame []ParseContext
	init  bool
}

func (c *ContextFrame) Init() {
	if c.init {
		return
	}
	c.frame = append(c.frame, TOP_LEVEL)
	c.init = true
}

func (c *ContextFrame) Begin(ctx ParseContext) {
	c.frame = append(c.frame, ctx)
}

func (c *ContextFrame) End() {
	c.frame = c.frame[:len(c.frame)-1]
}

func (c ContextFrame) Current() ParseContext {
	return c.frame[len(c.frame)-1]
}

// func (p *Parser) peek(n int) token.Token {
// 	for len(p.buffer) < n {
// 		p.buffer = append(p.buffer, p.lexer.Next())
// 	}

// 	return p.buffer[n-1]
// }

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
