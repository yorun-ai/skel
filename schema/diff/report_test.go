package diff_test

import (
	"testing"

	"go.yorun.ai/skel/schema/diff"
)

func TestReportHasImpact(t *testing.T) {
	report := new(diff.Report{Changes: []*diff.Change{
		new(diff.Change{Impact: diff.ImpactDangerous}),
	}})
	if !report.HasImpact(diff.ImpactDangerous) {
		t.Fatal("HasImpact(DANGEROUS) = false, want true")
	}
	if report.HasImpact(diff.ImpactBreaking) {
		t.Fatal("HasImpact(BREAKING) = true, want false")
	}
}
