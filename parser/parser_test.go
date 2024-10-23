package parser_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
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

func texpr(src string) ast.TypeExpr {
	expr, err := parser.ParseTypeExpr(file(src))
	if err != nil {
		panic(err)
	}
	return expr
}

type ExprCase struct {
	In       ast.Expr
	Out      ast.Expr
	Err      error
	Contains string
}

func (c *ExprCase) Run(assert *assert.Assertions) {
	fmt.Println(c.In, c.Out)
	assert.IsType(c.In, c.Out)

	s, _ := json.Marshal(c.In)
	fmt.Println(string(s))
	// out, err := checker.ResolveExpr(c.In)
	// if len(c.Contains) != 0 {
	// 	assert.ErrorContains(err, c.Contains)
	// }
	// if c.Err != nil {
	// 	assert.ErrorIs(wrap(err), c.Err)
	// } else {
	// 	assert.NoError(err)
	// }
	// assert.Equal(c.Out.Name(), out.Name())
}

func TestParser(t *testing.T) {
	tests := []Runner{
		&ExprCase{
			In:  expr("a > 0"),
			Out: &ast.BinaryExpr{
				// Operator: token.Token{Kind: token.GT},
				// Left:     &ast.IdentExpr{Value: "a"},
				// Right:    &ast.NumberLitExpr{Value: 0},
			},
		},
	}

	Run(tests, t)
}

func TestMarshal(t *testing.T) {
	assert := assert.New(t)

	file := file("declare i18n (en)")
	parser := parser.New(file)
	ast, err := parser.Parse()
	assert.NoError(err)
	b, err := json.MarshalIndent(ast, "", "  ")
	assert.NoError(err)
	assert.NotEmpty(b)

	fmt.Println(string(b))
}
