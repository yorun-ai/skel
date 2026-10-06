package schema

import (
	"encoding/json"
	"fmt"
	"io"
)

// Encode validates and writes one indented schema snapshot JSON document.
func Encode(writer io.Writer, document *Document) error {
	if err := Validate(document); err != nil {
		return err
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(document); err != nil {
		return fmt.Errorf("encode schema: %w", err)
	}
	return nil
}

// Decode reads exactly one schema snapshot JSON document, rejects unknown
// fields, and validates its format version and normalized structure.
func Decode(reader io.Reader) (*Document, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	document := new(Document)
	if err := decoder.Decode(document); err != nil {
		return nil, fmt.Errorf("decode schema: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode schema: unexpected trailing JSON value")
		}
		return nil, fmt.Errorf("decode schema: %w", err)
	}
	if err := Validate(document); err != nil {
		return nil, err
	}
	return document, nil
}
