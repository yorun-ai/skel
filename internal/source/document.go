// Package source owns immutable input revisions and byte-based source locations.
// Protocol-specific position encodings belong to their respective adapters.
package source

import (
	"crypto/sha256"
	"sort"
	"unicode/utf8"
)

// ID identifies a logical document independently of its optional physical path.
type ID string

// Span is a half-open byte range within a document revision.
type Span struct{ Start, End int }

// Document is immutable and safe to share between compiler and editor snapshots.
// Its digest and line index are computed once, when the revision is created.
type Document struct {
	id      ID
	path    string
	version int64
	text    string
	digest  [32]byte
	lines   []int
}

func New(id ID, path string, version int64, text string) *Document {
	lines := []int{0}
	for offset := range len(text) {
		if text[offset] == '\n' {
			lines = append(lines, offset+1)
		}
	}
	return new(Document{id: id, path: path, version: version, text: text, digest: sha256.Sum256([]byte(text)), lines: lines})
}
func (d *Document) ID() ID           { return d.id }
func (d *Document) Path() string     { return d.path }
func (d *Document) Version() int64   { return d.version }
func (d *Document) Text() string     { return d.text }
func (d *Document) Digest() [32]byte { return d.digest }
func (d *Document) LineCount() int   { return len(d.lines) }

// LineOffsets returns a zero-based line, excluding its newline and optional CR.
func (d *Document) LineOffsets(line int) (int, int, bool) {
	if line < 0 || line >= len(d.lines) {
		return 0, 0, false
	}
	start, end := d.lines[line], len(d.text)
	if line+1 < len(d.lines) {
		end = d.lines[line+1] - 1
	}
	if end > start && d.text[end-1] == '\r' {
		end--
	}
	return start, end, true
}
func (d *Document) Line(line int) string {
	start, end, ok := d.LineOffsets(line)
	if !ok {
		return ""
	}
	return d.text[start:end]
}

// LineAt returns the zero-based line and its byte offset, clamping at EOF.
func (d *Document) LineAt(offset int) (int, int) {
	offset = min(max(offset, 0), len(d.text))
	line := sort.Search(len(d.lines), func(i int) bool { return d.lines[i] > offset }) - 1
	return line, d.lines[line]
}

// Offset converts one-based parser line/rune columns to a byte offset.
func (d *Document) Offset(line, column int) int {
	start, end, ok := d.LineOffsets(max(line-1, 0))
	if !ok {
		return len(d.text)
	}
	for range max(column-1, 0) {
		if start >= end {
			break
		}
		_, width := utf8.DecodeRuneInString(d.text[start:end])
		start += width
	}
	return start
}
func (d *Document) Position(offset int) (line, column int) {
	offset = min(max(offset, 0), len(d.text))
	line, start := d.LineAt(offset)
	return line + 1, utf8.RuneCountInString(d.text[start:offset]) + 1
}
func (d *Document) IdentifierSpan(line, column int, name string) Span {
	start := d.Offset(line, column)
	return Span{Start: start, End: min(start+len(name), len(d.text))}
}

// AnalysisPath preserves physical diagnostic paths while retaining URI identities
// for virtual sources. Identity-sensitive caches must use ID instead.
func (d *Document) AnalysisPath() string {
	if d.path != "" {
		return d.path
	}
	return string(d.id)
}
