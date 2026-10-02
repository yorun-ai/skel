package workspace

import (
	"github.com/alecthomas/participle/v2/lexer"
	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/source"
)

func identifierRange(content source.Buffer, position lexer.Position, name string) protocol.Range {
	return content.IdentifierRange(position.Line, position.Column, name)
}
