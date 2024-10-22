package main

import (
	"fmt"
	"os"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
)

func FormatError(err error) string {
	tle, ok := err.(errs.TopLevelError)
	if !ok {
		return err.Error()
	}

	start, end := tle.Position()
	return fmt.Sprintf("%s at %s - %s in %s\n", tle.Error(), start.String(), end.String(), tle.File())
}

func main() {
	name := "test.lcl"
	r, _ := os.Open(name)
	file := parser.NewFile(name, r)

	p := parser.New(file)

	ast, err := p.Parse()
	if err != nil {
		fmt.Println(FormatError(err))
		os.Exit(1)
	}

	sc := analyzer.New(file, ast)
	err = sc.Scan()
	if err != nil {
		fmt.Println(FormatError(err))
		os.Exit(1)
	}
}
