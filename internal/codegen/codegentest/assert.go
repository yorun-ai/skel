package codegentest

import (
	"go/scanner"
	"go/token"
	"strings"
	"testing"
)

// AssertGoSourceContains reports a failure when src does not contain want as a
// Go token sequence.
//
// Generated files are gofmt-formatted. gofmt keeps the line breaks a generator
// emits and re-aligns keyed composite literal fields, so an assertion written
// against formatted text silently depends on that layout. Comparing token
// sequences keeps assertions about generated content independent of line
// breaks, indentation and column alignment.
func AssertGoSourceContains(t *testing.T, src string, want string) {
	t.Helper()
	if GoSourceContains(src, want) {
		return
	}
	t.Fatalf("generated source does not contain %s\ngot:\n%s", want, src)
}

// GoSourceContains reports whether src contains want as a Go token sequence.
func GoSourceContains(src string, want string) bool {
	return strings.Contains(GoSourceTokens(src), GoSourceTokens(want))
}

// GoSourceTokens renders the Go token stream of src as space separated tokens.
// String literal contents are preserved, and comments are kept as single
// tokens, so both stay comparable regardless of the surrounding layout.
func GoSourceTokens(src string) string {
	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(src))

	var lexer scanner.Scanner
	lexer.Init(file, []byte(src), func(token.Position, string) {}, 0)

	var tokens strings.Builder
	for {
		_, kind, literal := lexer.Scan()
		if kind == token.EOF {
			return tokens.String()
		}
		if kind == token.SEMICOLON && literal == "\n" {
			continue
		}
		if literal == "" {
			literal = kind.String()
		}
		if tokens.Len() > 0 {
			tokens.WriteByte(' ')
		}
		tokens.WriteString(literal)
	}
}
