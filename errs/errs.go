package errs

import (
	"reflect"
	"strings"

	"github.com/CanPacis/lcl/parser/token"
)

type ErrorSet struct {
	file   string
	Errors []error
}

func (e *ErrorSet) Error() string {
	return e.Errors[0].Error()
}

func (e *ErrorSet) Range() token.Range {
	p, ok := e.Errors[0].(interface {
		Range() token.Range
	})
	if ok {
		return p.Range()
	}

	return token.Range{}
}

func (e *ErrorSet) Unwrap() []error {
	return e.Errors
}

func (e *ErrorSet) File() string {
	return e.file
}

func NewErrorSet(file string, errors []error) error {
	return &ErrorSet{
		file:   file,
		Errors: errors,
	}
}

func join(elems any, sep string) string {
	rv := reflect.ValueOf(elems)

	list := []string{}
	for i := range rv.Len() {
		val := rv.Index(i).Elem()
		str, ok := val.Interface().(interface {
			String() string
		})
		if ok {
			list = append(list, str.String())
		}
	}
	return strings.Join(list, sep)
}
