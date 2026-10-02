package index

import (
	"fmt"
	"go.lsp.dev/uri"
	"strings"
	"testing"
)

func BenchmarkBuildDocument(b *testing.B) {
	for _, count := range []int{100, 1000, 4000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			var text strings.Builder
			text.WriteString("domain demo\n")
			for i := range count {
				fmt.Fprintf(&text, "data Value%d { next: Value%d? }\n", i, i)
			}
			content := text.String()
			b.ReportAllocs()
			b.SetBytes(int64(len(content)))
			b.ResetTimer()
			for b.Loop() {
				Build(uri.File("/audit/input.skel"), "/audit/input.skel", content, 1)
			}
		})
	}
}
