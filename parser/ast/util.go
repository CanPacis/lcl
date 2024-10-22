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

type stmtNode struct {
	Node
	comments []*CommentStmt
}

func (n *stmtNode) Comments() []*CommentStmt {
	return n.comments
}

func (n *stmtNode) stmtNode() {}

func NewStmtNode(s, e token.Position, comments ...*CommentStmt) *stmtNode {
	return &stmtNode{
		Node:     NewNode(s, e),
		comments: comments,
	}
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
