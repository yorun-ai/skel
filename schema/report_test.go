package schema_test

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestReportHasImpact(t *testing.T) {
	report := new(schema.Report{Changes: []*schema.Change{
		new(schema.Change{Impact: schema.ImpactDangerous}),
	}})
	if !report.HasImpact(schema.ImpactDangerous) {
		t.Fatal("HasImpact(DANGEROUS) = false, want true")
	}
	if report.HasImpact(schema.ImpactBreaking) {
		t.Fatal("HasImpact(BREAKING) = true, want false")
	}
}
