package parser_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CanPacis/go-i18n/parser"
)

func file(src string) *parser.File {
	return parser.NewFile("test.lcl", bytes.NewBuffer([]byte(src)))
}

func TestParser(t *testing.T) {
	file := file(`for (en) in i18n
		fn(r::Range) Duration Time.Between(r.start r.end)
	`)
	parser := parser.New(file)
	ast, err := parser.Parse()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ast)
}
