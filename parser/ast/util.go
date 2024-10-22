package ast

import "github.com/CanPacis/go-i18n/parser/token"

type node struct {
	start, end token.Position
}

func (n *node) Start() token.Position {
	return n.start
}

func (n *node) End() token.Position {
	return n.end
}

func NewNode(s, e token.Position) *node {
	return &node{start: s, end: e}
}

func NewIdent(t token.Token) *IdentExpr {
	return &IdentExpr{
		Node:  NewNode(t.Start, t.End),
		Value: t.Literal,
	}
}

func NewString(t token.Token) *StringLitExpr {
	return &StringLitExpr{
		Node:  NewNode(t.Start, t.End),
		Value: t.Literal,
	}
}
