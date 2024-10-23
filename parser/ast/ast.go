package ast

import (
	"github.com/CanPacis/go-i18n/parser/token"
)

type Node interface {
	NodeType() string

	Start() token.Position
	End() token.Position
}

const FileNode = "file"

type File struct {
	Node    `json:"node"`
	Decl    *DeclStmt     `json:"decl"`
	Imports []*ImportStmt `json:"imports"`
	Stmts   []Stmt        `json:"stmts"`
}

type Stmt interface {
	Node
	Comments() []*CommentStmt
	stmtNode()
}

const DeclTargetNode = "decl_target"

type DeclTarget struct {
	Node `json:"node"`
	Tag  *StringLitExpr `json:"tag"`
	Name *IdentExpr     `json:"name"`
}

const DeclStmtNode = "decl_stmt"

type DeclStmt struct {
	Stmt    `json:"node"`
	Name    *IdentExpr    `json:"name"`
	Targets []*DeclTarget `json:"targets"`
}

const ImportStmtNode = "import_stmt"

type ImportStmt struct {
	Stmt `json:"node"`
	List []*IdentExpr `json:"list"`
}

const TypeDefStmtNode = "type_def_stmt"

type TypeDefStmt struct {
	Stmt `json:"node"`
	Name *IdentExpr `json:"name"`
	Type TypeExpr   `json:"type"`
}

const ParameterNode = "parameter"

type Parameter struct {
	Node  `json:"node"`
	Index int        `json:"index"`
	Name  *IdentExpr `json:"name"`
	Type  TypeExpr   `json:"type"`
}

const FnDefStmtNode = "fn_def_stmt"

type FnDefStmt struct {
	Stmt   `json:"node"`
	Name   *IdentExpr   `json:"name"`
	Params []*Parameter `json:"params"`
	Body   Expr         `json:"body"`
}

const SectionStmtNode = "section_stmt"

type SectionStmt struct {
	Stmt `json:"node"`
	Name *IdentExpr `json:"name"`
	Body []Entry    `json:"body"`
}

const CommentStmtNode = "comment_stmt"

type CommentStmt struct {
	Stmt    `json:"node"`
	Literal string `json:"literal"`
	Raw     string `json:"raw"`
}

const EmptyStmtNode = "empty_stmt"

type EmptyStmt struct {
	Stmt `json:"node"`
}

type Entry interface {
	Node
	entryNode()
}

const KeyEntryNode = "key_entry"

type KeyEntry struct {
	Node   `json:"node"`
	Name   *IdentExpr `json:"name"`
	Fields []*Field   `json:"fields"`
}

const TemplateEntryNode = "template_entry"

type TemplateEntry struct {
	Node        `json:"node"`
	Partitioned bool         `json:"partitioned"`
	Name        *IdentExpr   `json:"name"`
	Fields      []*Field     `json:"fields"`
	Params      []*Parameter `json:"params"`
}

const EmptyEntryNode = "empty_entry"

type EmptyEntry struct {
	Node `json:"node"`
}

func (e *SectionStmt) entryNode()   {}
func (e *KeyEntry) entryNode()      {}
func (e *TemplateEntry) entryNode() {}
func (e *EmptyEntry) entryNode()    {}

const FieldNode = "field"

type Field struct {
	Node  `json:"node"`
	Tag   *IdentExpr `json:"tag"`
	Value Expr       `json:"value"`
}

type Expr interface {
	Node
	exprNode()
}

type TypeExpr interface {
	Node
	tExprNode()
}

const BinaryExprNode = "binary_expr"

type BinaryExpr struct {
	Node     `json:"node"`
	Operator token.Token `json:"operator"`
	Left     Expr        `json:"left"`
	Right    Expr        `json:"right"`
}

const ArithmeticExprNode = "arithmetic_expr"

type ArithmeticExpr struct {
	Node     `json:"node"`
	Operator token.Token `json:"operator"`
	Left     Expr        `json:"left"`
	Right    Expr        `json:"right"`
}

const TernaryExprNode = "ternary_expr"

type TernaryExpr struct {
	Node      `json:"node"`
	Predicate Expr `json:"predicate"`
	Left      Expr `json:"left"`
	Right     Expr `json:"right"`
}

const CallExprNode = "call_expr"

type CallExpr struct {
	Node `json:"node"`
	Fn   Expr   `json:"fn"`
	Args []Expr `json:"args"`
}

const MemberExprNode = "member_expr"

type MemberExpr struct {
	Node  `json:"node"`
	Left  Expr       `json:"left"`
	Right *IdentExpr `json:"right"`
}

const IndexExprNode = "index_expr"

type IndexExpr struct {
	Node  `json:"node"`
	Host  Expr           `json:"host"`
	Index *NumberLitExpr `json:"index"`
}

const GroupExprNode = "group_expr"

type GroupExpr struct {
	Node `json:"node"`
	Expr Expr `json:"expr"`
}

const IdentExprNode = "ident_expr"

type IdentExpr struct {
	Node  `json:"node"`
	Value string `json:"value"`
}

const StringLitExprNode = "string_literal_expr"

type StringLitExpr struct {
	Node  `json:"node"`
	Value string `json:"value"`
}

const TemplateLitExprNode = "template_literal_expr"

type TemplateLitExpr struct {
	Node  `json:"node"`
	Value []Expr `json:"value"`
}

const NumberLitExprNode = "number_literal_expr"

type NumberLitExpr struct {
	Node  `json:"node"`
	Value float64 `json:"value"`
}

const EmptyExprNode = "empty_expr"

type EmptyExpr struct {
	Node `json:"node"`
}

const TypeMemberExprNode = "type_member_expr"

type TypeMemberExpr struct {
	Node  `json:"node"`
	Left  *IdentExpr `json:"left"`
	Right *IdentExpr `json:"right"`
}

const ListTypeExprNode = "list_type_expr"

type ListTypeExpr struct {
	Node `json:"node"`
	Type TypeExpr `json:"type"`
}

const StructLitExprNode = "struct_literal_expr"

type StructLitExpr struct {
	Node `json:"node"`
	List []*TypePair `json:"list"`
}

const TypePairNode = "type_pair"

type TypePair struct {
	Node  `json:"node"`
	Index int        `json:"index"`
	Name  *IdentExpr `json:"name"`
	Type  TypeExpr   `json:"type"`
}

func (e *BinaryExpr) exprNode()      {}
func (e *ArithmeticExpr) exprNode()  {}
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
