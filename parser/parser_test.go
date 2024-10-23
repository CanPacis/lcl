package parser_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
)

func FormatError(err error) string {
	tle, ok := err.(errs.TopLevelError)
	if !ok {
		return err.Error()
	}

	start, end := tle.Position()
	return fmt.Sprintf("%s at %s - %s in %s", tle.Error(), start, end, tle.File())
}

var Test = map[string]*parser.File{}

func init() {
	raw, _ := os.ReadFile("test.lcl")
	sections := strings.Split(string(raw), "#test:")

	for _, entry := range sections {
		if len(entry) == 0 {
			continue
		}

		split := strings.Split(entry, "\n")
		name := strings.TrimSpace(split[0])
		content := strings.Join(split[1:], "\n")
		Test[name] = parser.NewFile(fmt.Sprintf("%s.lcl", name), bytes.NewBuffer([]byte(content)))
	}
}

func TestParser(t *testing.T) {
	start := time.Now()
	parser := parser.New(Test["Section0"])
	f, err := parser.Parse()
	fmt.Println(time.Since(start))
	if err != nil {
		fmt.Println(FormatError(err))
	}
	body := f.Stmts[0].(*ast.FnDefStmt).Body
	lit := body.(*ast.TemplateLitExpr)

	for _, expr := range lit.Value {
		fmt.Println(expr, reflect.TypeOf(expr))
	}
	tern := lit.Value[3].(*ast.TernaryExpr)
	fmt.Println(tern.Predicate)
	fmt.Println(tern.Left)
	fmt.Println(tern.Right)
}
