package analyzer_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/CanPacis/go-i18n/errs"
	"github.com/CanPacis/go-i18n/parser"
	"github.com/stretchr/testify/assert"
)

type err struct {
	err error
}

func (e *err) Error() string {
	if e == nil {
		return ""
	}
	return "wrapped: " + e.err.Error()
}

func (e *err) Unwrap() error {
	return e.err
}

func (e *err) Is(target error) bool {
	return reflect.TypeOf(e.err).String() == reflect.TypeOf(target).String()
}

func wrap(e error) error {
	if e == nil {
		return nil
	}
	return &err{err: e}
}

type Runner interface {
	Run(*assert.Assertions)
}

func file(src string) *parser.File {
	return parser.NewFile("test.lcl", bytes.NewBuffer([]byte(src)))
}

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
