package diff_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/codegen"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

func inspectSchema(t *testing.T, path, source string) *schema.Domain {
	t.Helper()
	result, err := api.QuerySchema(api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(source)}}, api.SchemaQueryOption{})
	if err != nil {
		t.Fatal(err)
	}
	return result.Domain
}

func TestCompareRecursiveSchemaUsesReferenceIdentity(t *testing.T) {
	source := `domain demo
 data Node { value: string next: Node? }
 data Box<TItem> { item: TItem }
 service NodesService { method read { output Box<Node> } }
 `
	before := inspectSchema(t, filepath.Join(t.TempDir(), "before.skel"), source)
	after := inspectSchema(t, filepath.Join(t.TempDir(), "after.skel"), "\n\n"+source)
	// Derived hashes and object addresses are not source contract changes.
	after.Data()[0].Hash = "different hash"
	report, err := diff.Compare(before, after)
	if err != nil || len(report.Changes) != 0 {
		t.Fatalf("same semantic graph changed: %+v, %v", report, err)
	}
	changed := inspectSchema(t, filepath.Join(t.TempDir(), "changed.skel"), strings.Replace(source, "value: string", "value: int", 1))
	report, err = diff.Compare(before, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Code != "data.member.type.changed" || report.Changes[0].Symbol != "demo.Node.value" {
		t.Fatalf("reference expanded into spurious changes: %+v", report.Changes)
	}
	change := report.Changes[0]
	if change.Baseline == nil || change.Candidate == nil || change.Baseline.File == change.Candidate.File {
		t.Fatalf("lost independent source positions: %+v", change)
	}
}

func TestCompareUnresolvedSchemaCarriesImportIdentity(t *testing.T) {
	source := `domain demo
 import foreign.contract as imported
 pub data Envelope { values: list<imported.Box<imported.Value>> }
 pub service ReadsService { require imported.Record:read:owns(id) method read { input { id: string } output Envelope } }
 `
	before := inspectSchema(t, filepath.Join(t.TempDir(), "before.skel"), source)
	after := inspectSchema(t, filepath.Join(t.TempDir(), "after.skel"), strings.ReplaceAll(source, "imported", "renamed"))
	if len(before.Imports()) != 1 || before.Imports()[0].Name != "foreign.contract" || before.Imports()[0].Domain != nil {
		t.Fatalf("lost unresolved import: %+v", before.Imports())
	}
	ref := before.Data()[0].Members[0].Type.List.Value
	if before.TypeReferenceName(ref) != "foreign.contract.Box" || before.TypeReferenceName(ref.TypeArguments[0]) != "foreign.contract.Value" {
		t.Fatalf("lost external identity: %+v", ref)
	}
	if _, err := codegen.Prepare(before, codegen.Selection{}); err == nil {
		t.Fatal("unresolved schema accepted for code generation")
	}
	report, err := diff.Compare(before, after)
	if err != nil || len(report.Changes) != 0 {
		t.Fatalf("alias affected comparison: %+v, %v", report, err)
	}
	if before.Imports()[0].Alias != "imported" || after.Imports()[0].Alias != "renamed" {
		t.Fatal("comparison mutated source aliases")
	}
}

func TestCompareRejectsMissingDomains(t *testing.T) {
	domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo"})
	for _, pair := range [][2]*schema.Domain{{nil, domain}, {domain, nil}, {nil, nil}} {
		if _, err := diff.Compare(pair[0], pair[1]); err == nil {
			t.Fatal("missing domain accepted")
		}
	}
}
