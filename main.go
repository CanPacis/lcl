package main

import (
	"fmt"
	"os"

	"github.com/CanPacis/go-i18n/analyzer"
	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
)

func main() {
	name := "test.lcl"
	r, _ := os.Open(name)
	file := parser.NewFile(name, r)

	p := parser.New(file)

	ast, err := p.ParseFile()
	if err != nil {
		e := err.(*errs.SyntaxError)
		start, end := e.Position()
		fmt.Printf("%s at %s - %s in %s\n", e.Error(), start.String(), end.String(), e.File)
		os.Exit(1)
	}

	anly := analyzer.New(file, ast)
	err = anly.Run()
	if err != nil {
		e := err.(*errs.SemanticError)
		start, end := e.Position()
		fmt.Printf("%s at %s - %s in %s\n", e.Error(), start.String(), end.String(), e.File)
		os.Exit(1)
	}

	exports := anly.Scope.Exports()
	fmt.Println(exports["test"].Name())
	// fmt.Println(anly.TypeEnv.Lookup("User", ""))
	// fmt.Println(anly.TypeEnv.Exports())
	// fmt.Println(anly.Scope.Lookup("string", ""))
	// fmt.Println(anly.Scope.Exports())
}
