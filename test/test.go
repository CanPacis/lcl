package test

import (
	"bytes"
	"testing"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/CanPacis/go-i18n/parser/ast"
	"github.com/stretchr/testify/assert"
)

type Runner interface {
	Run(*assert.Assertions)
}

func Run(t *testing.T, cases []Runner) {
	assert := assert.New(t)
	for _, c := range cases {
		c.Run(assert)
	}
}

type Injector[D any] interface {
	Runner
	Inject(D)
}

func RunWith[D any](t *testing.T, cases []Injector[D], dep D) {
	assert := assert.New(t)
	for _, c := range cases {
		c.Inject(dep)
		c.Run(assert)
	}
}

type ParserOptions struct {
	File *parser.File
}

func WithFile(file *parser.File) func(*ParserOptions) {
	return func(po *ParserOptions) {
		po.File = file
	}
}

func WithName(name string) func(*ParserOptions) {
	return func(po *ParserOptions) {
		po.File.Name = name
	}
}

func WithSourceString(source string) func(*ParserOptions) {
	return func(po *ParserOptions) {
		po.File = parser.NewFile(po.File.Name, bytes.NewBuffer([]byte(source)))
	}
}

func WithSourceBytes(source []byte) func(*ParserOptions) {
	return func(po *ParserOptions) {
		po.File = parser.NewFile(po.File.Name, bytes.NewBuffer(source))
	}
}

func NewParser(options ...func(*ParserOptions)) *parser.Parser {
	opts := &ParserOptions{
		File: &parser.File{
			Name: "mock.lcl",
		},
	}

	for _, option := range options {
		option(opts)
	}

	return parser.New(opts.File)
}

func Parse(options ...func(*ParserOptions)) (*ast.File, error) {
	parser := NewParser(options...)
	return parser.Parse()
}

func MustParse(options ...func(*ParserOptions)) *ast.File {
	file, err := Parse(options...)
	if err != nil {
		panic(FormatError(err))
	}
	return file
}

func ParseExpr(options ...func(*ParserOptions)) (ast.Expr, error) {
	opts := &ParserOptions{
		File: &parser.File{
			Name: "mock.lcl",
		},
	}

	for _, option := range options {
		option(opts)
	}

	return parser.ParseExpr(opts.File)
}

func MustParseExpr(options ...func(*ParserOptions)) ast.Expr {
	expr, err := ParseExpr(options...)
	if err != nil {
		panic(FormatError(err))
	}
	return expr
}

func ParseTypeExpr(options ...func(*ParserOptions)) (ast.TypeExpr, error) {
	opts := &ParserOptions{
		File: &parser.File{
			Name: "mock.lcl",
		},
	}

	for _, option := range options {
		option(opts)
	}

	return parser.ParseTypeExpr(opts.File)
}

func MustParseTypeExpr(options ...func(*ParserOptions)) ast.TypeExpr {
	expr, err := ParseTypeExpr(options...)
	if err != nil {
		panic(FormatError(err))
	}
	return expr
}

func FormatError(err error) string {
	es, ok := err.(*errs.ErrorSet)
	if !ok {
		return err.Error()
	}

	start, end := es.Position()
	pos := start.String() + " - " + end.String()
	return es.Error() + " at " + pos + " in " + es.File()
}
