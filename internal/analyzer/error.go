package analyzer

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

type MissingImportError struct {
	Position schema.Position
	Domain   string
}

func (e *MissingImportError) Error() string {
	return fmt.Sprintf("%s skel import %s not found; pass --skel-import %s=PATH", e.Position, e.Domain, e.Domain)
}

func (e *MissingImportError) SourcePosition() schema.Position { return e.Position }
