package gogen_test

import (
	"bytes"
	goast "go/ast"
	goparser "go/parser"
	"go/printer"
	gotoken "go/token"
	"testing"

	gogen "github.com/CanPacis/lcl/gen/go"
	"github.com/CanPacis/lcl/ir"
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/parser/token"
	"github.com/CanPacis/lcl/test"
	"github.com/CanPacis/lcl/types"
	"github.com/stretchr/testify/assert"
)

type ExprCase struct {
	In  ast.Expr
	Out string
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	out := gogen.ResolveExpr(c.In)

	fset := gotoken.NewFileSet()
	buf := bytes.NewBuffer([]byte{})
	printer.Fprint(buf, fset, out)
	// Maybe parse the expexted go output too?
	assert.Equal(c.Out, buf.String())
}

type TypeExprCase struct {
	In  types.Type
	Out string
}

func (c *TypeExprCase) Run(assert *assert.Assertions) {
	out := gogen.ResolveTypeExpr(c.In)

	fset := gotoken.NewFileSet()
	buf := bytes.NewBuffer([]byte{})
	printer.Fprint(buf, fset, out)
	// Maybe parse the expexted go output too?
	assert.Equal(c.Out, buf.String())
}

type FuncDeclCase struct {
	In   *ir.FnDef
	Recv string
	Out  string
}

func (c *FuncDeclCase) Run(assert *assert.Assertions) {
	out := gogen.GenerateFuncDecl(c.In, c.Recv)

	fset := gotoken.NewFileSet()

	actual := bytes.NewBuffer([]byte{})
	printer.Fprint(actual, fset, out)

	ast, err := goparser.ParseExpr(c.Out)
	assert.NoError(err)
	lit := ast.(*goast.FuncLit)
	expected := bytes.NewBuffer([]byte{})
	printer.Fprint(expected, fset, &goast.FuncDecl{
		Name: out.Name,
		Type: lit.Type,
		Body: lit.Body,
	})

	assert.Equal(expected.String(), actual.String())
}

func TestUtil(t *testing.T) {
	fn := &ir.FnDef{
		Definition: ir.NewDefinition("Test", true),
		Stmt: &ast.FnDefStmt{
			Name: &ast.IdentExpr{Value: "Test"},
			Params: []*ast.TypePair{
				{
					Index: 0,
					Name:  &ast.IdentExpr{Value: "ages"},
					Type: &ast.ListTypeExpr{
						Type: &ast.IdentExpr{Value: "int"},
					},
				},
			},
			Body: &ast.TernaryExpr{
				Predicate: &ast.BinaryExpr{
					Operator: token.Token{Kind: token.GT},
					Left: &ast.IndexExpr{
						Host:  &ast.IdentExpr{Value: "ages"},
						Index: &ast.NumberLitExpr{Value: 0},
					},
					Right: &ast.NumberLitExpr{Value: 18},
				},
				Left:  &ast.IdentExpr{Value: "true"},
				Right: &ast.IdentExpr{Value: "false"},
			},
		},
		Type: &types.Fn{
			In:  []types.Type{types.NewList(types.Int)},
			Out: types.Bool,
		},
	}

	tests := []test.Runner{
		&ExprCase{
			In:  &ast.IdentExpr{Value: "test"},
			Out: "test",
		},
		// &GenExprCase{
		// 	In:  &ast.StringLitExpr{Value: "test"},
		// 	Out: `"test"`,
		// },
		&ExprCase{
			In:  &ast.NumberLitExpr{Value: 5},
			Out: "5",
		},
		&ExprCase{
			In:  &ast.NumberLitExpr{Value: 5.2},
			Out: "5.200000",
		},
		&ExprCase{
			In:  &ast.GroupExpr{Expr: &ast.NumberLitExpr{Value: 0}},
			Out: "(0)",
		},
		&ExprCase{
			In: &ast.MemberExpr{
				Left:  &ast.IdentExpr{Value: "user"},
				Right: &ast.IdentExpr{Value: "age"},
			},
			Out: "user.age",
		},
		&ExprCase{
			In: &ast.ArithmeticExpr{
				Operator: token.Token{Kind: token.PLUS},
				Left:     &ast.NumberLitExpr{Value: 8},
				Right:    &ast.IdentExpr{Value: "count"},
			},
			Out: "8 + count",
		},
		&ExprCase{
			In: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.GTE},
				Left:     &ast.NumberLitExpr{Value: 8},
				Right:    &ast.NumberLitExpr{Value: 9},
			},
			Out: "8 >= 9",
		},
		&ExprCase{
			In: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.LTE},
				Left:     &ast.NumberLitExpr{Value: 8.1},
				Right:    &ast.NumberLitExpr{Value: 9.1},
			},
			Out: "8.100000 <= 9.100000",
		},
		&ExprCase{
			In: &ast.CallExpr{
				Fn:   &ast.IdentExpr{Value: "fn"},
				Args: []ast.Expr{},
			},
			Out: "fn()",
		},
		&ExprCase{
			In: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "fn"},
				Args: []ast.Expr{
					&ast.IdentExpr{Value: "i"},
					&ast.IdentExpr{Value: "j"},
				},
			},
			Out: "fn(i, j)",
		},
		&TypeExprCase{
			In:  types.Bool,
			Out: "bool",
		},
		&TypeExprCase{
			In:  types.Int,
			Out: "int",
		},
		&TypeExprCase{
			In:  types.U16,
			Out: "uint16",
		},
		&TypeExprCase{
			In:  types.F64,
			Out: "float64",
		},
		&TypeExprCase{
			In:  types.NewList(types.Byte),
			Out: "[]byte",
		},
		&TypeExprCase{
			In:  types.NewList(types.NewList(types.Rune)),
			Out: "[][]rune",
		},
		&FuncDeclCase{
			In:  fn,
			Out: `func(ages []int) bool { return tern(ages[0] > 18, true, false) }`,
		},
	}

	test.Run(t, tests)
}
