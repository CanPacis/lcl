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
	Decl    *DeclStmt
	Imports []*ImportStmt
	Stmts   []Stmt
}

type Stmt interface {
	Node
	Comments() []*CommentStmt
	stmtNode()
}

type DeclStmt struct {
	Stmt
	Name *IdentExpr
	List []*IdentExpr
}

type ImportStmt struct {
	Stmt
	List []*IdentExpr
}

type TypeDefStmt struct {
	Stmt
	Name *IdentExpr
	Type TypeExpr
}

type Parameter struct {
	Node
	Index int
	Name  *IdentExpr
	Type  TypeExpr
}

type FnDefStmt struct {
	Stmt
	Name   *IdentExpr
	Params []*Parameter
	Body   Expr
}

type SectionStmt struct {
	Stmt
	Name *IdentExpr
	Body []Entry
}

type CommentStmt struct {
	Stmt
	Literal string
	Raw     string
}

type EmptyStmt struct {
	Stmt
}

type Entry interface {
	Node
	entryNode()
}

type KeyEntry struct {
	Node
	Name   *IdentExpr
	Fields []Field
}

type TemplateEntry struct {
	Node
	Partitioned bool
	Name        *IdentExpr
	Fields      []Field
	// Type        TypeExpr
	Params []*Parameter
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

type CallExpr struct {
	Node
	Fn   Expr
	Args []Expr
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
	Index int
	Name  *IdentExpr
	Type  TypeExpr
}

func (e *BinaryExpr) exprNode()      {}
func (e *TernaryExpr) exprNode()     {}
func (e *CallExpr) exprNode()        {}
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
