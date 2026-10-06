package analyzer

import (
	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/schema"
)

func position(pos lexer.Position) schema.Position {
	return schema.Position{File: pos.Filename, Line: pos.Line, Column: pos.Column}
}
