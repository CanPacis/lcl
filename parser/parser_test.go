package parser_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/CanPacis/go-i18n/parser/token"
	"github.com/stretchr/testify/assert"
)

func FormatError(err error) string {
	tle, ok := err.(errs.TopLevelError)
	if !ok {
		return err.Error()
	}

	start, end := tle.Position()
	return fmt.Sprintf("%s at %s - %s in %s", tle.Error(), start, end, tle.File())
}

type Runner interface {
	Run(*assert.Assertions)
}

func Run(cases []Runner, t *testing.T) {
	assert := assert.New(t)

	for _, c := range cases {
		c.Run(assert)
	}
}

func file(src string) *parser.File {
	return parser.NewFile("test.lcl", bytes.NewBuffer([]byte(src)))
}

func expr(src string) ast.Expr {
	expr, err := parser.ParseExpr(file(src))
	if err != nil {
		panic(err)
	}
	return expr
}

// func texpr(src string) ast.TypeExpr {
// 	expr, err := parser.ParseTypeExpr(file(src))
// 	if err != nil {
// 		panic(err)
// 	}
// 	return expr
// }

type ExprCase struct {
	In       string
	Out      ast.Expr
	Err      error
	Contains string
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	expr, err := parser.ParseExpr(file(c.In))
	if c.Out != nil {
		CompareExpr(A{assert}, c.Out, expr)
	} else {
		if len(c.Contains) > 0 {
			assert.ErrorContains(err, c.Contains)
		} else {
			if err != nil {
				assert.Fail(FormatError(err))
			}
		}
	}
}

func TestExpr(t *testing.T) {
	tests := []Runner{
		&ExprCase{
			In: "ident",
			Out: &ast.IdentExpr{
				Value: "ident",
			},
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
		&ExprCase{In: `""()`, Contains: errs.Unexpected},
	}

	Run(tests, t)
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	file := file(`declare i18n ("en-US" as en)`)
	parser := parser.New(file)
	ast, err := parser.Parse()
	assert.NoError(err)
	b, err := json.MarshalIndent(ast, "", "  ")
	assert.NoError(err)
	assert.NotEmpty(b)
}
