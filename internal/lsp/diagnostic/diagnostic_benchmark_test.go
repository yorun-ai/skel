package diagnostic

import (
	skeldiagnostic "go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/lsp/source"
	"go.yorun.ai/skel/schema"
	"strings"
	"testing"
)

func BenchmarkDiagnosticRange(b *testing.B) {
	text := strings.Repeat("data User { id: int }\n", 10000)
	r := skeldiagnostic.SourceRange{Start: schema.Position{Line: 9999, Column: 6}, End: schema.Position{Line: 9999, Column: 10}}
	b.Run("build-buffer", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Range(text, r)
		}
	})
	buffer := source.New(text)
	b.Run("reuse-buffer", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			RangeBuffer(buffer, r)
		}
	})
}
