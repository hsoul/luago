package lexer

import (
	"fmt"
	"os"
	"regexp"
	"testing"
)

func kindToCategory(kind int) string {
	switch {
	case kind < TOKEN_SEP_SEMI:
		return "other"
	case kind <= TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= TOKEN_OP_NOT:
		return "operator"
	case kind <= TOKEN_KW_WHILE:
		return "keyword"
	case kind == TOKEN_IDENTIFIER:
		return "identifier"
	case kind == TOKEN_NUMBER:
		return "number"
	case kind == TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}

// go test -v -run TestLexer //  go test -v
func TestLexer(t *testing.T) {
	data, err := os.ReadFile("../../test.lua")
	if err != nil {
		panic(err)
	}
	lexer := NewLexer(string(data), "test")
	for {
		line, kind, token := lexer.NextToken()
		fmt.Printf("[%2d] [%-10s] %s\n", line, kindToCategory(kind), token)
		if kind == TOKEN_EOF {
			break
		}
	}
}

// var regNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
// var regIdentifier = regexp.MustCompile(`^[_\d\w]+`)
// var regNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)
// var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)
var regOpeningLongBracket = regexp.MustCompile(`^\[=*\[`)

// var regDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
// var regHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
// var regUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

func TestRegexp(t *testing.T) {
	str := "[["
	fmt.Println(regOpeningLongBracket.MatchString(str))
}
