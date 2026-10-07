package location

import "fmt"

// Position identifies a one-based location in a Skel source file.
// A zero Position is unknown.
type Position struct {
	// File is the source path and may be empty when no path is available.
	File string `json:"file,omitempty"`
	// Line is the one-based source line.
	Line int `json:"line"`
	// Column is the one-based source column.
	Column int `json:"column"`
}

// String formats p as file:line:column, or line:column when File is empty.
func (p Position) String() string {
	if p.File == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}
