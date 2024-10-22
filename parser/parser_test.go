package parser_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/CanPacis/go-i18n/parser"
)

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
	parser.Parse()
	fmt.Println(time.Since(start))
}
