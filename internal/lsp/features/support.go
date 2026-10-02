package features

import (
	"slices"
	"strings"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/source"
	"go.yorun.ai/skelc/internal/lsp/workspace"
)

func utf16Length(value string) int {
	return source.UTF16Length(value)
}

func isIdentifierValue(value string) bool {
	return workspace.IsIdentifier(value)
}

func containsPosition(r protocol.Range, position protocol.Position) bool {
	return source.ContainsPosition(r, position)
}

func sortLocations(locations []protocol.Location) {
	slices.SortFunc(locations, func(left, right protocol.Location) int {
		if compared := strings.Compare(string(left.URI), string(right.URI)); compared != 0 {
			return compared
		}
		return source.ComparePosition(left.Range.Start, right.Range.Start)
	})
}
