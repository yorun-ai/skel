package compiler

import (
	"strings"
	"unicode/utf8"

	"go.yorun.ai/skel/internal/model"
	textsource "go.yorun.ai/skel/internal/source"
)

func sourceLineOffsets(source []byte, line int) (int, int, bool) {
	return textsource.New("", "", 0, string(source)).LineOffsets(line - 1)
}

func sourceRangeAt(start model.Position, source []byte) SourceRange {
	return sourceRangeAtDocument(start, textsource.New("", "", 0, string(source)))
}

func sourceRangeAtDocument(start model.Position, document *textsource.Document) SourceRange {
	end := start
	if document == nil || start.Line <= 0 || start.Column <= 0 {
		return SourceRange{Start: start, End: end}
	}
	lineStart, lineEnd, ok := document.LineOffsets(start.Line - 1)
	if !ok {
		return SourceRange{Start: start, End: end}
	}
	source := document.Text()
	offset := lineStart
	for range start.Column - 1 {
		if offset >= lineEnd {
			break
		}
		_, width := utf8.DecodeRuneInString(source[offset:lineEnd])
		offset += width
	}
	for offset < lineEnd && (source[offset] == ' ' || source[offset] == '\t') {
		offset++
	}
	endOffset := offset
	if endOffset < lineEnd && isSourceIdentifierByte(source[endOffset]) {
		for endOffset < lineEnd && isSourceIdentifierByte(source[endOffset]) {
			endOffset++
		}
		for endOffset+1 < lineEnd && source[endOffset] == '.' && isSourceIdentifierByte(source[endOffset+1]) {
			endOffset++
			for endOffset < lineEnd && isSourceIdentifierByte(source[endOffset]) {
				endOffset++
			}
		}
	} else {
		for endOffset < lineEnd {
			value, width := utf8.DecodeRuneInString(source[endOffset:lineEnd])
			if strings.ContainsRune(" \t,.:;(){}[]<>?=@", value) {
				break
			}
			endOffset += width
		}
	}
	if endOffset == offset && endOffset < lineEnd {
		_, width := utf8.DecodeRuneInString(source[endOffset:lineEnd])
		endOffset += width
	}
	end.Column = 1 + utf8.RuneCountInString(source[lineStart:endOffset])
	if end.Column <= start.Column {
		end.Column = start.Column + 1
	}
	return SourceRange{Start: start, End: end}
}

func isSourceIdentifierByte(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
