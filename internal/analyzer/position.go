package analyzer

import (
	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/model"
)

func position(pos lexer.Position) model.Position {
	return model.Position{File: pos.Filename, Line: pos.Line, Column: pos.Column}
}
