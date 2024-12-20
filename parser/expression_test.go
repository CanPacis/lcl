package parser_test

import (
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/stretchr/testify/assert"
)

// to make func signatures precise
type A struct {
	*assert.Assertions
}

func CompareExpr(assert A, left, right ast.Expr) {
	switch left := left.(type) {
	case *ast.BinaryExpr:
		assert.IsType(left, right)
		r := right.(*ast.BinaryExpr)

		assert.Equal(left.Operator.Kind, r.Operator.Kind)
		CompareExpr(assert, left.Left, r.Left)
		CompareExpr(assert, left.Right, r.Right)
	case *ast.ArithmeticExpr:
		r := right.(*ast.ArithmeticExpr)

		assert.Equal(left.Operator.Kind, r.Operator.Kind)
		CompareExpr(assert, left.Left, r.Left)
		CompareExpr(assert, left.Right, r.Right)
	case *ast.TernaryExpr:
		assert.IsType(left, right)
		r := right.(*ast.TernaryExpr)

		CompareExpr(assert, left.Predicate, r.Predicate)
		CompareExpr(assert, left.Left, r.Left)
		CompareExpr(assert, left.Right, r.Right)
	case *ast.CallExpr:
		assert.IsType(left, right)
		r := right.(*ast.CallExpr)

		CompareExpr(assert, left.Fn, r.Fn)
		assert.Equal(len(left.Args), len(r.Args), "invalid number of args are passed to call")
		for i, arg := range left.Args {
			CompareExpr(assert, arg, r.Args[i])
		}
	case *ast.MemberExpr:
		assert.IsType(left, right)
		r := right.(*ast.MemberExpr)

		CompareExpr(assert, left.Left, r.Left)
		CompareIdent(assert, left.Right, r.Right)
	case *ast.ImportExpr:
		assert.IsType(left, right)
		r := right.(*ast.ImportExpr)

		CompareIdent(assert, left.Left, r.Left)
		CompareIdent(assert, left.Right, r.Right)
	case *ast.IndexExpr:
		assert.IsType(left, right)
		r := right.(*ast.IndexExpr)

		CompareExpr(assert, left.Host, r.Host)
		CompareExpr(assert, left.Index, r.Host)
	case *ast.GroupExpr:
		assert.IsType(left, right)
		r := right.(*ast.GroupExpr)
		CompareExpr(assert, left.Expr, r.Expr)
	case *ast.IdentExpr:
		CompareIdent(assert, left, right)
	case *ast.StringLitExpr:
		CompareString(assert, left, right)
	case *ast.TemplateLitExpr:
		CompareTemplate(assert, left, right)
	case *ast.NumberLitExpr:
		CompareNumber(assert, left, right)
	case *ast.EmptyExpr:
		assert.Fail("Left hand side is empty", left)
	}
}

func CompareIdent(assert A, left *ast.IdentExpr, right ast.Expr) {
	assert.IsType(left, right)
	r := right.(*ast.IdentExpr)
	assert.Equal(left.Value, r.Value)
}

func CompareString(assert A, left *ast.StringLitExpr, right ast.Expr) {
	assert.IsType(left, right)
	r := right.(*ast.StringLitExpr)
	assert.Equal(left.Value, r.Value)
}

func CompareTemplate(assert A, left *ast.TemplateLitExpr, right ast.Expr) {
	assert.IsType(left, right)
	r := right.(*ast.TemplateLitExpr)

	assert.Equal(len(left.Value), len(r.Value), "Template has invalid number of expressions")

	for i, expr := range left.Value {
		CompareExpr(assert, expr, r.Value[i])
	}
}

func CompareNumber(assert A, left *ast.NumberLitExpr, right ast.Expr) {
	assert.IsType(left, right)
	r := right.(*ast.NumberLitExpr)
	assert.Equal(left.Value, r.Value)
}

func CompareTypeExpr(assert A, left ast.TypeExpr, right ast.TypeExpr) {
	switch left := left.(type) {
	case *ast.IdentExpr:
		assert.IsType(left, right)
		r := right.(*ast.IdentExpr)
		assert.Equal(left.Value, r.Value)
	case *ast.ImportExpr:
		assert.IsType(left, right)
		r := right.(*ast.ImportExpr)

		CompareIdent(assert, left.Left, r.Left)
		CompareIdent(assert, left.Right, r.Right)
	case *ast.ListTypeExpr:
		assert.IsType(left, right)
		r := right.(*ast.ListTypeExpr)
		CompareTypeExpr(assert, left.Type, r.Type)
	case *ast.StructLitExpr:
		assert.IsType(left, right)
		r := right.(*ast.StructLitExpr)

		assert.Equal(len(left.Fields), len(r.Fields), "Struct literal has invalid number of fields")

		for i, typ := range left.Fields {
			CompareTypeExpr(assert, typ.Type, r.Fields[i].Type)
		}
	case *ast.EmptyExpr:
		assert.Fail("Left hand side is empty", left)
	}
}
