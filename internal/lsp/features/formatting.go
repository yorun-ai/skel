package features

import (
	"context"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skel/internal/formatter"
)

func (s *Service) Formatting(_ context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	document := s.Snapshot.Document(params.TextDocument.URI)
	if document == nil || len(document.ParseDiagnostics) > 0 {
		return []protocol.TextEdit{}, nil
	}
	content, err := formatter.Source([]byte(document.Source))
	if err != nil {
		return nil, err
	}
	formatted := string(content)
	if formatted == document.Source {
		return []protocol.TextEdit{}, nil
	}
	return []protocol.TextEdit{{
		Range:   protocol.Range{Start: protocol.Position{}, End: document.Buffer.Position(len(document.Source))},
		NewText: formatted,
	}}, nil
}
