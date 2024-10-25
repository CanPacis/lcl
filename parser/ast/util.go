package ast

import (
	"encoding/json"

	"github.com/CanPacis/lcl/parser/token"
)

type node struct {
	typ        string
	start, end token.Position
}

func (n *node) NodeType() string {
	return n.typ
}

func (n *node) Start() token.Position {
	return n.start
}

func (n *node) End() token.Position {
	return n.end
}

func (n *node) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":  n.NodeType(),
		"start": n.Start(),
		"end":   n.End(),
	})
}

func NewNode(t string, s, e token.Position) *node {
	return &node{typ: t, start: s, end: e}
}

type stmtNode struct {
	Node
	comments []*CommentStmt
}

func (n *stmtNode) Comments() []*CommentStmt {
	return n.comments
}

func (n *stmtNode) stmtNode() {}

func (n *stmtNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"comments": n.Comments(),
		"type":     n.NodeType(),
		"start":    n.Start(),
		"end":      n.End(),
	})
}

func NewStmtNode(t string, s, e token.Position, comments ...*CommentStmt) *stmtNode {
	c := []*CommentStmt{}
	if comments != nil {
		c = comments
	}

	return &stmtNode{
		Node:     NewNode(t, s, e),
		comments: c,
	}
}

func NewIdent(t token.Token) *IdentExpr {
	return &IdentExpr{
		Node:  NewNode(IdentExprNode, t.Start, t.End),
		Value: t.Literal,
	}
}

func NewString(t token.Token) *StringLitExpr {
	return &StringLitExpr{
		Node:  NewNode(StringLitExprNode, t.Start, t.End),
		Value: t.Literal,
	}
}
