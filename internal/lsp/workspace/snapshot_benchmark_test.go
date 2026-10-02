package workspace

import (
	"fmt"
	"testing"

	"go.lsp.dev/uri"
)

func BenchmarkSnapshot(b *testing.B) {
	for _, count := range []int{100, 1000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			store := New()
			for i := range count {
				store.Put(uri.File(fmt.Sprintf("/audit/%d.skel", i)), fmt.Sprintf("domain demo.d%d\ndata User { next: User? }\n", i), 1, true)
			}
			store.Snapshot()
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				store.Snapshot()
			}
		})
	}
}

func BenchmarkSnapshotAfterEdit(b *testing.B) {
	store := New()
	for i := range 1000 {
		store.Put(uri.File(fmt.Sprintf("/audit/%d.skel", i)), fmt.Sprintf("domain demo.d%d\ndata User {}\n", i), 1, true)
	}
	documentURI := uri.File("/audit/0.skel")
	version := int32(1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		version++
		store.Put(documentURI, fmt.Sprintf("domain demo.d0\ndata Value%d {}\n", version), version, true)
		store.Snapshot()
	}
}
