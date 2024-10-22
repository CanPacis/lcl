package ast

import (
	"github.com/CanPacis/go-i18n/parser/token"
)

type Node interface {
	Start() token.Position
	End() token.Position
}

type File struct {
	Node
	For     *ForStmt
	Imports []*ImportStmt
	Stmts   []Stmt
}

type Stmt interface {
	Node
	stmtNode()
}

type ForStmt struct {
	Node
	List []*IdentExpr
	In   *IdentExpr
}

type ImportStmt struct {
	Node
	List []*IdentExpr
}

type TypeDefStmt struct {
	Node
	Name *IdentExpr
	Type TypeExpr
}

type ProcDefStmt struct {
	Node
	Name *IdentExpr
	Body Expr
}

type SectionStmt struct {
	Node
	Name *IdentExpr
	Body []Entry
}

type CommentStmt struct {
	Node
	Literal string
	Raw     string
}

type EmptyStmt struct {
	Node
}

func (s *ForStmt) stmtNode()     {}
func (s *ImportStmt) stmtNode()  {}
func (s *TypeDefStmt) stmtNode() {}
func (s *ProcDefStmt) stmtNode() {}
func (s *SectionStmt) stmtNode() {}
func (s *CommentStmt) stmtNode() {}
func (s *EmptyStmt) stmtNode()   {}

type Entry interface {
	Node
	entryNode()
}

type KeyEntry struct {
	Node
	Name   *IdentExpr
	Fields []*StringField
}

type TemplateEntry struct {
	Node
	Partitioned bool
	Name        *IdentExpr
	Fields      []*TemplateField
	Type        TypeExpr
}

type EmptyEntry struct {
	Node
}

func (e *SectionStmt) entryNode()   {}
func (e *KeyEntry) entryNode()      {}
func (e *TemplateEntry) entryNode() {}
func (e *EmptyEntry) entryNode()    {}

type Field interface {
	Node
	Target() string
}

type StringField struct {
	Node
	Tag   *IdentExpr
	Value *StringLitExpr
}

func (f *StringField) Target() string {
	return f.Tag.Value
}

type TemplateField struct {
	Node
	Tag   *IdentExpr
	Value *TemplateLitExpr
}

func (f *TemplateField) Target() string {
	return f.Tag.Value
}

type Expr interface {
	Node
	exprNode()
}

type TypeExpr interface {
	Node
	tExprNode()
}

type SelfExpr struct {
	Node
}

type BinaryExpr struct {
	Node
	Operator token.Token
	Left     Expr
	Right    Expr
}

type TernaryExpr struct {
	Node
	Predicate Expr
	Left      Expr
	Right     Expr
}

type ProcCallExpr struct {
	Node
	Proc  Expr
	Param Expr
}

type MemberExpr struct {
	Node
	Left  Expr
	Right *IdentExpr
}

type IndexExpr struct {
	Node
	Host  Expr
	Index *NumberLitExpr
}

type GroupExpr struct {
	Node
	Expr Expr
}

type IdentExpr struct {
	Node
	Value string
}

type StringLitExpr struct {
	Node
	Value string
}

type TemplateLitExpr struct {
	Node
	Value []Expr
}

type NumberLitExpr struct {
	Node
	Value float64
}

type EmptyExpr struct {
	Node
}

type TypeMemberExpr struct {
	Node
	Left  *IdentExpr
	Right *IdentExpr
}

type ListTypeExpr struct {
	Node
	Type TypeExpr
}

type StructLitExpr struct {
	Node
	List []*TypePair
}

type TypePair struct {
	Node
	Name *IdentExpr
	Type TypeExpr
}

func (e *SelfExpr) exprNode()        {}
func (e *BinaryExpr) exprNode()      {}
func (e *TernaryExpr) exprNode()     {}
func (e *ProcCallExpr) exprNode()    {}
func (e *MemberExpr) exprNode()      {}
func (e *IndexExpr) exprNode()       {}
func (e *GroupExpr) exprNode()       {}
func (e *IdentExpr) exprNode()       {}
func (e *StringLitExpr) exprNode()   {}
func (e *TemplateLitExpr) exprNode() {}
func (e *NumberLitExpr) exprNode()   {}
func (e *EmptyExpr) exprNode()       {}

func (e *IdentExpr) tExprNode()      {}
func (e *TypeMemberExpr) tExprNode() {}
func (e *ListTypeExpr) tExprNode()   {}
func (e *StructLitExpr) tExprNode()  {}
func (e *EmptyExpr) tExprNode()      {}
