package analyzer

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

type MissingImportError struct {
	Position model.Position
	Domain   string
}

func (e *MissingImportError) Error() string {
	return fmt.Sprintf("%s skel import %s not found; pass --skel-import %s=PATH", e.Position, e.Domain, e.Domain)
}

func (e *MissingImportError) SourcePosition() model.Position { return e.Position }
