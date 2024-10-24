package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
	"github.com/CanPacis/go-i18n/test"
	"github.com/stretchr/testify/assert"
)

type ExprCase struct {
	In       string
	Out      ast.Expr
	Err      error
	Contains string
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	expr, err := test.ParseExpr(test.WithSourceString(c.In))

	if c.Out != nil {
		CompareExpr(A{assert}, c.Out, expr)
	} else {
		if c.Err != nil {
			assert.ErrorIs(err, c.Err)
		} else {
			assert.NoError(err)
		}
	}
}

func TestExpr(t *testing.T) {
	tests := []test.Runner{
		&ExprCase{
			In:  "3",
			Out: &ast.NumberLitExpr{Value: 3},
		},
		&ExprCase{
			In:  "3.1",
			Out: &ast.NumberLitExpr{Value: 3.1},
		},
		&ExprCase{
			In:  "-3.1",
			Out: &ast.NumberLitExpr{Value: -3.1},
		},
		&ExprCase{
			In:  "-0",
			Out: &ast.NumberLitExpr{Value: 0},
		},
		&ExprCase{
			In:  `""`,
			Out: &ast.StringLitExpr{Value: ""},
		},
		&ExprCase{
			In:  `"literal"`,
			Out: &ast.StringLitExpr{Value: "literal"},
		},
		&ExprCase{
			In:  `"literal`,
			Err: errs.ErrUnterminatedConstruct,
		},
		&ExprCase{
			In: "``",
			Out: &ast.TemplateLitExpr{
				Value: []ast.Expr{
					&ast.StringLitExpr{Value: ""},
				},
			},
		},
		&ExprCase{
			In: "`basic`",
			Out: &ast.TemplateLitExpr{
				Value: []ast.Expr{
					&ast.StringLitExpr{Value: "basic"},
				},
			},
		},
		&ExprCase{
			In: "`has { expressions } inside { call() }`",
			Out: &ast.TemplateLitExpr{
				Value: []ast.Expr{
					&ast.StringLitExpr{Value: "has "},
					&ast.IdentExpr{Value: "expressions"},
					&ast.StringLitExpr{Value: " inside "},
					&ast.CallExpr{
						Fn:   &ast.IdentExpr{Value: "call"},
						Args: []ast.Expr{},
					},
					&ast.StringLitExpr{Value: ""},
				},
			},
		},
		&ExprCase{
			In:  "`unterminated template ",
			Err: errs.ErrUnterminatedConstruct,
		},
		&ExprCase{
			In:  "`unterminated { expression `",
			Err: errs.ErrUnterminatedConstruct,
		},
		&ExprCase{In: "`{}{\n}\n{}`"},
		&ExprCase{
			In:  "ident",
			Out: &ast.IdentExpr{Value: "ident"},
		},
		&ExprCase{
			In: "ident < 0",
			Out: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.LT},
				Left:     &ast.IdentExpr{Value: "ident"},
				Right:    &ast.NumberLitExpr{Value: 0},
			},
		},
		&ExprCase{
			In: "ident >= 0",
			Out: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.GTE},
				Left:     &ast.IdentExpr{Value: "ident"},
				Right:    &ast.NumberLitExpr{Value: 0},
			},
		},
		&ExprCase{
			In: "ident == ident",
			Out: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.EQUALS},
				Left:     &ast.IdentExpr{Value: "ident"},
				Right:    &ast.IdentExpr{Value: "ident"},
			},
		},
		&ExprCase{
			In: "3 + 7",
			Out: &ast.ArithmeticExpr{
				Operator: token.Token{Kind: token.PLUS},
				Left:     &ast.NumberLitExpr{Value: 3},
				Right:    &ast.NumberLitExpr{Value: 7},
			},
		},
		&ExprCase{
			In: "3 + 7 * 5",
			Out: &ast.ArithmeticExpr{
				Operator: token.Token{Kind: token.PLUS},
				Left:     &ast.NumberLitExpr{Value: 3},
				Right: &ast.ArithmeticExpr{
					Operator: token.Token{Kind: token.STAR},
					Left:     &ast.NumberLitExpr{Value: 7},
					Right:    &ast.NumberLitExpr{Value: 5},
				},
			},
		},
		&ExprCase{
			In: "3 + 7 * 5 ^ 2",
			Out: &ast.ArithmeticExpr{
				Operator: token.Token{Kind: token.PLUS},
				Left:     &ast.NumberLitExpr{Value: 3},
				Right: &ast.ArithmeticExpr{
					Operator: token.Token{Kind: token.STAR},
					Left:     &ast.NumberLitExpr{Value: 7},
					Right: &ast.ArithmeticExpr{
						Operator: token.Token{Kind: token.CARET},
						Left:     &ast.NumberLitExpr{Value: 5},
						Right:    &ast.NumberLitExpr{Value: 2},
					},
				},
			},
		},
		&ExprCase{In: `""()`, Err: errs.ErrUnexpectedToken},
		&ExprCase{In: `member.of`},
		&ExprCase{In: `member.of.long`},
		&ExprCase{In: `member.of.long()`},
		&ExprCase{
			In: `imports::package`,
			Out: &ast.ImportExpr{
				Left:  &ast.IdentExpr{Value: "imports"},
				Right: &ast.IdentExpr{Value: "package"},
			},
		},
		&ExprCase{
			In: `imports::func()`,
			Out: &ast.CallExpr{
				Fn: &ast.ImportExpr{
					Left:  &ast.IdentExpr{Value: "imports"},
					Right: &ast.IdentExpr{Value: "func"},
				},
			},
		},
		&ExprCase{
			In: `imports::member.of`,
			Out: &ast.MemberExpr{
				Left: &ast.ImportExpr{
					Left:  &ast.IdentExpr{Value: "imports"},
					Right: &ast.IdentExpr{Value: "member"},
				},
				Right: &ast.IdentExpr{Value: "of"},
			},
		},
		&ExprCase{
			In: `imports::member.of.complex`,
			Out: &ast.MemberExpr{
				Left: &ast.MemberExpr{
					Left: &ast.ImportExpr{
						Left:  &ast.IdentExpr{Value: "imports"},
						Right: &ast.IdentExpr{Value: "member"},
					},
					Right: &ast.IdentExpr{Value: "of"},
				},
				Right: &ast.IdentExpr{Value: "complex"},
			},
		},
		&ExprCase{In: `imports::member.of.complex + -32`},
		&ExprCase{In: `imports::member.of.complex + -32 == 0`},
		&ExprCase{In: `(imports::member.of.complex + -32 == 0) || false`},
		&ExprCase{In: `index[0]`},
		&ExprCase{In: `index.of[0]`},
		&ExprCase{In: `index.of[identifier]`},
		&ExprCase{In: `index.of[call()]`},
		&ExprCase{In: `index.of[call(imports::member.of.complex + -32 == 0)]`},
		&ExprCase{In: `index.of[call(imports::member.of.complex + -32 == 0) extra]`, Err: errs.ErrUnexpectedToken},
		&ExprCase{In: `call(param)`},
		&ExprCase{
			In: `call(param1 param2)`,
			Out: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "call"},
				Args: []ast.Expr{
					&ast.IdentExpr{Value: "param1"},
					&ast.IdentExpr{Value: "param2"},
				},
			},
		},
		&ExprCase{In: `call(param.of index[0])`},
		&ExprCase{In: `call(param.of (a || b) 6)`},
		&ExprCase{In: `call()()`},
		&ExprCase{In: `call(param.of a || (b 6))`, Err: errs.ErrUnexpectedToken},
		&ExprCase{
			In: `call(param.of a || b 6)`,
			Out: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "call"},
				Args: []ast.Expr{
					&ast.MemberExpr{
						Left:  &ast.IdentExpr{Value: "param"},
						Right: &ast.IdentExpr{Value: "of"},
					},
					&ast.BinaryExpr{
						Operator: token.Token{Kind: token.OR},
						Left:     &ast.IdentExpr{Value: "a"},
						Right:    &ast.IdentExpr{Value: "b"},
					},
					&ast.NumberLitExpr{Value: 6},
				},
			},
		},
		&ExprCase{In: `pred ? if : else`},
		&ExprCase{In: `pred ? a == b : a || c`},
		&ExprCase{In: `call() ? a == b : a || c`},
		&ExprCase{
			In: `call(true ? a == b : -30.1 || m.of it::continues)`,
			Out: &ast.CallExpr{
				Fn: &ast.IdentExpr{Value: "call"},
				Args: []ast.Expr{
					&ast.TernaryExpr{
						Predicate: &ast.IdentExpr{Value: "true"},
						Left: &ast.BinaryExpr{
							Operator: token.Token{Kind: token.EQUALS},
							Left:     &ast.IdentExpr{Value: "a"},
							Right:    &ast.IdentExpr{Value: "b"},
						},
						Right: &ast.BinaryExpr{
							Operator: token.Token{Kind: token.OR},
							Left:     &ast.NumberLitExpr{Value: -30.1},
							Right: &ast.MemberExpr{
								Left:  &ast.IdentExpr{Value: "m"},
								Right: &ast.IdentExpr{Value: "of"},
							},
						},
					},
					&ast.ImportExpr{
						Left:  &ast.IdentExpr{Value: "it"},
						Right: &ast.IdentExpr{Value: "continues"},
					},
				},
			},
		},
		&ExprCase{
			In: `3 + 5 == true`,
			Out: &ast.BinaryExpr{
				Operator: token.Token{Kind: token.EQUALS},
				Left: &ast.ArithmeticExpr{
					Operator: token.Token{Kind: token.PLUS},
					Left:     &ast.NumberLitExpr{Value: 3},
					Right:    &ast.NumberLitExpr{Value: 5},
				},
				Right: &ast.IdentExpr{Value: "true"},
			},
		},
		&ExprCase{In: "`template { call(true ? a == b : -30.1 || m.of it::continues) } with { complex.expressions[0] }`"},
	}

	test.Run(t, tests)
}

type TypeExprCase struct {
	In       string
	Out      ast.TypeExpr
	Err      error
	Contains string
}

func (c *TypeExprCase) Run(assert *assert.Assertions) {
	expr, err := test.ParseTypeExpr(test.WithSourceString(c.In))

	if c.Out != nil {
		CompareTypeExpr(A{assert}, c.Out, expr)
	} else {
		if len(c.Contains) > 0 {
			assert.ErrorContains(err, c.Contains)
		} else {
			if err != nil {
				assert.Fail(test.FormatError(err))
			}
		}
	}
}

func TestTypeExpr(t *testing.T) {
	tests := []test.Runner{
		&TypeExprCase{
			In:  "string",
			Out: &ast.IdentExpr{Value: "string"},
		},
		&TypeExprCase{
			In: "int[]",
			Out: &ast.ListTypeExpr{
				Type: &ast.IdentExpr{Value: "int"},
			},
		},
		&TypeExprCase{
			In: "time::Time",
			Out: &ast.ImportExpr{
				Left:  &ast.IdentExpr{Value: "time"},
				Right: &ast.IdentExpr{Value: "Time"},
			},
		},
		&TypeExprCase{
			In:  "{}",
			Out: &ast.StructLitExpr{Fields: []*ast.TypePair{}},
		},
		&TypeExprCase{
			In: "time::Time[]",
			Out: &ast.ListTypeExpr{
				Type: &ast.ImportExpr{
					Left:  &ast.IdentExpr{Value: "time"},
					Right: &ast.IdentExpr{Value: "Time"},
				},
			},
		},
		&TypeExprCase{
			In: "time::Time[][]",
			Out: &ast.ListTypeExpr{
				Type: &ast.ListTypeExpr{
					Type: &ast.ImportExpr{
						Left:  &ast.IdentExpr{Value: "time"},
						Right: &ast.IdentExpr{Value: "Time"},
					},
				},
			},
		},
	}
	test.Run(t, tests)
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	ast, err := test.Parse(test.WithSourceString(`declare i18n ("en-US" as en_us)`))
	assert.NoError(err)
	b, err := json.MarshalIndent(ast, "", "  ")
	assert.NoError(err)
	assert.NotEmpty(b)
}
