package formatter

import (
	"bytes"
	"fmt"
	"strings"
)

// Source returns the canonical representation of one Skel source file.
// Formatting is intentionally syntax preserving: declarations are never
// reordered, and multiline string or comment indentation is only rebased.
func Source(source []byte) ([]byte, error) {
	normalized := bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
	tokens, err := lex(normalized)
	if err != nil {
		return nil, fmt.Errorf("lex Skel source: %w", err)
	}

	var audiences []_Audience
	boundaries := sourceBoundaries(normalized, &audiences)
	tokens, continuations := layoutAudiences(tokens, audiences, boundaries)
	layout := buildLayout(tokens, boundaries, continuations)
	return render(layout), nil
}

func finishSource(source []byte) []byte {
	trimmed := strings.TrimRight(string(source), " \t\n")
	if trimmed == "" {
		return nil
	}
	return []byte(trimmed + "\n")
}
