package formatter

import "go.yorun.ai/skel/internal/parser"

// ValidatedSource validates syntax before and after canonical formatting.
func ValidatedSource(content []byte) ([]byte, error) {
	if err := parser.ValidateSource("<source>", content); err != nil {
		return nil, err
	}
	formatted, err := Source(content)
	if err != nil {
		return nil, err
	}
	if err := parser.ValidateSource("<source>", formatted); err != nil {
		return nil, err
	}
	return formatted, nil
}
