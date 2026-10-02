package index

import (
	"strings"
	"unicode"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/source"
)

func indexIncompleteDocument(document *Document, tokens []source.Token) {
	depth := 0
	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		switch token.Value {
		case "{":
			depth++
			continue
		case "}":
			depth = max(depth-1, 0)
			continue
		}
		if depth != 0 {
			continue
		}
		switch token.Value {
		case "domain":
			name, _ := qualifiedTokenValue(tokens, index+1)
			if name != "" {
				document.Domain = name
			}
		case "import":
			domain, next := qualifiedTokenValue(tokens, index+1)
			if domain == "" {
				continue
			}
			alias := domain
			if next+1 < len(tokens) && tokens[next].Value == "as" && IsIdentifier(tokens[next+1].Value) {
				alias = tokens[next+1].Value
			}
			document.Imports[alias] = domain
		default:
			kind, detail := fallbackEntryKind(token.Value)
			if kind == protocol.SymbolKindNull || index+1 >= len(tokens) || !IsIdentifier(tokens[index+1].Value) {
				continue
			}
			nameToken := tokens[index+1]
			range_ := document.Buffer.Range(nameToken.Start, nameToken.End)
			document.Definitions = append(document.Definitions, Definition{
				Key: document.Domain + "." + nameToken.Value, Name: nameToken.Value, Detail: detail, Kind: kind, Range: range_,
			})
			document.Symbols = append(document.Symbols, Symbol{Name: nameToken.Value, Detail: detail, Kind: kind, Range: range_})
		}
	}
}

func qualifiedTokenValue(tokens []source.Token, start int) (string, int) {
	if start >= len(tokens) || !IsIdentifier(tokens[start].Value) {
		return "", start
	}
	parts := []string{tokens[start].Value}
	index := start + 1
	for index+1 < len(tokens) && tokens[index].Value == "." && IsIdentifier(tokens[index+1].Value) {
		parts = append(parts, tokens[index+1].Value)
		index += 2
	}
	return strings.Join(parts, "."), index
}

func fallbackEntryKind(keyword string) (protocol.SymbolKind, string) {
	switch keyword {
	case "enum":
		return protocol.SymbolKindEnum, keyword
	case "data", "config":
		return protocol.SymbolKindStruct, keyword
	case "actor", "service", "web":
		return protocol.SymbolKindInterface, keyword
	case "resource":
		return protocol.SymbolKindObject, keyword
	case "event":
		return protocol.SymbolKindEvent, keyword
	case "task":
		return protocol.SymbolKindFunction, keyword
	default:
		return protocol.SymbolKindNull, ""
	}
}

// IsIdentifier reports whether value is a valid Skel identifier.
func IsIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for index, r := range value {
		if index == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
		} else if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
