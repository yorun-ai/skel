package parser

import (
	"bytes"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

// SourceSegment is one independently recoverable top-level source fragment.
type SourceSegment struct {
	Start int
	End   int
	Line  int
}

type _SourceSegmentScanner struct {
	source     []byte
	tokens     []lexer.Token
	identifier lexer.TokenType
	elided     map[lexer.TokenType]bool
	starts     []SourceSegment

	depth                 int
	hasDeclaration        bool
	lastLine              int
	pendingDecoratorStart int
	pendingDecoratorLine  int
}

// SplitSourceSegments identifies top-level fragments without applying compiler
// recovery or diagnostic policy.
func SplitSourceSegments(path string, source []byte) []SourceSegment {
	lex, err := grammar.LexerDefinition().Lex(path, bytes.NewReader(source))
	if err != nil {
		return []SourceSegment{{End: len(source), Line: 1}}
	}
	tokens, err := lexer.ConsumeAll(lex)
	if err != nil {
		return []SourceSegment{{End: len(source), Line: 1}}
	}

	symbols := grammar.LexerDefinition().Symbols()
	elided := map[lexer.TokenType]bool{
		symbols["Whitespace"]:   true,
		symbols["LineComment"]:  true,
		symbols["BlockComment"]: true,
		symbols["Newline"]:      true,
	}
	scanner := &_SourceSegmentScanner{
		source: source, tokens: tokens, identifier: symbols["Identifier"], elided: elided,
		starts: []SourceSegment{{Line: 1}}, pendingDecoratorStart: -1,
	}
	return scanner.scan()
}

func (s *_SourceSegmentScanner) scan() []SourceSegment {
	for index, token := range s.tokens {
		s.consume(index, token)
	}
	for index := range s.starts {
		if index+1 < len(s.starts) {
			s.starts[index].End = s.starts[index+1].Start
		} else {
			s.starts[index].End = len(s.source)
		}
	}
	return s.starts
}

func (s *_SourceSegmentScanner) consume(index int, token lexer.Token) {
	if token.EOF() || s.elided[token.Type] {
		return
	}
	// A decorator after a declaration on the same line belongs to the next
	// declaration. Keep both in one fragment so its original column survives.
	if token.Pos.Line == s.lastLine && token.Value == "@" && s.depth == 0 {
		s.hasDeclaration = false
	}
	if token.Pos.Line != s.lastLine {
		s.consumeFirstTokenOnLine(index, token)
		s.lastLine = token.Pos.Line
	}
	s.updateDepth(token.Value)
}

func (s *_SourceSegmentScanner) consumeFirstTokenOnLine(index int, token lexer.Token) {
	kind := topLevelTokenKind(s.tokens, index, s.identifier, s.elided)
	if s.depth > 0 && kind == "decorator" {
		if s.pendingDecoratorStart < 0 {
			s.pendingDecoratorStart, _, _ = segmentLineOffsets(s.source, token.Pos.Line)
			s.pendingDecoratorLine = token.Pos.Line
		}
		return
	}
	if kind == "" {
		s.pendingDecoratorStart = -1
		return
	}
	if s.hasDeclaration || s.depth > 0 {
		start, _, _ := segmentLineOffsets(s.source, token.Pos.Line)
		line := token.Pos.Line
		if s.depth > 0 && s.pendingDecoratorStart >= 0 {
			start = s.pendingDecoratorStart
			line = s.pendingDecoratorLine
		}
		s.starts = append(s.starts, SourceSegment{Start: start, Line: line})
	}
	s.hasDeclaration = kind != "decorator"
	if s.depth > 0 {
		s.depth = 0
	}
	s.pendingDecoratorStart = -1
}

func segmentLineOffsets(source []byte, line int) (int, int, bool) {
	if line <= 0 {
		return 0, 0, false
	}
	start := 0
	for current := 1; current < line; current++ {
		index := bytes.IndexByte(source[start:], '\n')
		if index < 0 {
			return 0, 0, false
		}
		start += index + 1
	}
	end := len(source)
	if index := bytes.IndexByte(source[start:], '\n'); index >= 0 {
		end = start + index
	}
	return start, end, true
}

func (s *_SourceSegmentScanner) updateDepth(value string) {
	switch value {
	case "{":
		s.depth++
	case "}":
		if s.depth > 0 {
			s.depth--
		}
	}
}

func topLevelTokenKind(tokens []lexer.Token, index int, identifier lexer.TokenType, elided map[lexer.TokenType]bool) string {
	value := tokens[index].Value
	if value == "@" {
		return "decorator"
	}
	line := tokens[index].Pos.Line
	for value == "pub" || value == "api" || value == "open" {
		index = nextSignificantToken(tokens, index+1, line, elided)
		if index < 0 {
			return ""
		}
		value = tokens[index].Value
	}
	keywords := append([]string{"domain", "import"}, grammar.TopLevelDeclarationKeywords()...)
	for _, keyword := range keywords {
		next := nextSignificantToken(tokens, index+1, line, elided)
		if value == keyword && next >= 0 && tokens[next].Type == identifier {
			return keyword
		}
	}
	return ""
}

func nextSignificantToken(tokens []lexer.Token, index, line int, elided map[lexer.TokenType]bool) int {
	for ; index < len(tokens); index++ {
		if tokens[index].EOF() || tokens[index].Pos.Line != line {
			return -1
		}
		if !elided[tokens[index].Type] {
			return index
		}
	}
	return -1
}
