package formatter

import (
	"bytes"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
)

type _Token struct {
	kind   string
	value  string
	column int
}

func lex(source []byte) ([]_Token, error) {
	definition := grammar.LexerDefinition()
	symbols := lexer.SymbolsByRune(definition)
	stream, err := definition.Lex("", bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	tokens := make([]_Token, 0, len(source)/4)
	for {
		token, err := stream.Next()
		if err != nil {
			return nil, err
		}
		if token.EOF() {
			return tokens, nil
		}
		kind := symbols[token.Type]
		if kind == "Whitespace" {
			continue
		}
		tokens = append(tokens, _Token{kind: kind, value: token.Value, column: token.Pos.Column})
	}
}
