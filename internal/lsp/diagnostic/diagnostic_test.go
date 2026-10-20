package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	skeldiagnostic "go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
)

func TestToProtocolConvertsUTF16RangeAndRelatedInformation(t *testing.T) {
	item := skeldiagnostic.Diagnostic{
		Code: skeldiagnostic.CodeSemanticDuplicate, Severity: skeldiagnostic.SeverityError,
		Range: skeldiagnostic.SourceRange{
			Start: schema.Position{Line: 1, Column: 2},
			End:   schema.Position{Line: 1, Column: 3},
		},
		Message: "duplicate",
		Related: []skeldiagnostic.RelatedInformation{{
			Range:   skeldiagnostic.SourceRange{Start: schema.Position{File: "/workspace/first.skel", Line: 2, Column: 1}},
			Message: "first declaration",
		}},
	}

	converted := ToProtocol(item, "😀x", func(path string) (uri.URI, string, bool) {
		return uri.File(path), "domain demo", true
	})

	assert.Equal(t, protocol.Position{Character: 2}, converted.Range.Start)
	assert.Equal(t, protocol.Position{Character: 3}, converted.Range.End)
	require.Len(t, converted.RelatedInformation, 1)
	assert.Equal(t, uri.File("/workspace/first.skel"), converted.RelatedInformation[0].Location.URI)
}
