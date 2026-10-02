package features

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"go.lsp.dev/protocol"
	lsource "go.yorun.ai/skelc/internal/lsp/source"
)

func qualifierBeforePositionBuffer(buffer lsource.Buffer, position protocol.Position) string {
	source := buffer.String()
	offset := buffer.Offset(position)
	start := offset
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(source[:start])
		if r != '_' && !isLetterOrDigit(r) {
			break
		}
		start -= size
	}
	if start == 0 || source[start-1] != '.' {
		return ""
	}
	end := start - 1
	start = end
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(source[:start])
		if r != '.' && r != '_' && !isLetterOrDigit(r) {
			break
		}
		start -= size
	}
	return source[start:end]
}

func decoratorPrefixBeforePositionBuffer(buffer lsource.Buffer, position protocol.Position) (string, protocol.Range, bool) {
	source := buffer.String()
	offset := buffer.Offset(position)
	start := offset
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(source[:start])
		if r != '_' && !isLetterOrDigit(r) {
			break
		}
		start -= size
	}
	if start == 0 || source[start-1] != '@' {
		return "", protocol.Range{}, false
	}
	return source[start:offset], buffer.Range(start, offset), true
}

func completionValuesBeforePositionBuffer(buffer lsource.Buffer, position protocol.Position) []string {
	source := buffer.String()
	offset := buffer.Offset(position)
	lineStart := strings.LastIndexByte(source[:offset], '\n') + 1
	prefix := strings.TrimSpace(source[lineStart:offset])
	fields := strings.Fields(prefix)
	if len(fields) == 0 {
		return nil
	}
	if fields[0] == "pub" || fields[0] == "api" || fields[0] == "open" {
		fields = fields[1:]
		if len(fields) == 0 {
			return nil
		}
	}
	trailingSpace := offset > lineStart && (source[offset-1] == ' ' || source[offset-1] == '\t')
	if len(fields) >= 2 && fields[0] == "config" {
		switch {
		case len(fields) == 2 && trailingSpace:
			return configLifecycleCompletionValues
		case len(fields) == 3 && isIdentifierValue(fields[2]):
			return configLifecycleCompletionValues
		}
	}
	last := len(fields) - 1
	switch {
	case fields[last] == "via" && trailingSpace:
		return actorViaCompletionValues
	case last > 0 && fields[last-1] == "via" && isIdentifierValue(fields[last]):
		return actorViaCompletionValues
	}
	return nil
}

func isLetterOrDigit(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
