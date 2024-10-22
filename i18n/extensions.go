package i18n

import (
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
)

type time_ext struct{}

func (e time_ext) Year(t time.Time) int {
	return t.Year()
}

var time_ext_i = time_ext{}

type strconv_ext struct{}

func (e strconv_ext) Itoa(i int) string {
	return strconv.Itoa(i)
}

var strconv_ext_i = strconv_ext{}

type tr_case_ext struct{}

func (e tr_case_ext) Ablative(s string) string {
	return "den"
}

var tr_case_ext_i = tr_case_ext{}

type list_ext struct {
	tag language.Tag
}

func (e list_ext) CommaJoin(s []string) string {
	return strings.Join(s, ", ")
}

var list_ext_en = list_ext{tag: language.English}
var list_ext_tr = list_ext{tag: language.Turkish}
