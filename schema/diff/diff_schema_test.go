package diff_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/codegen"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

func inspectImportedSchema(t *testing.T, source string, resolved bool) *schema.Domain {
	t.Helper()
	root := t.TempDir()
	path, imported := filepath.Join(root, "source.skel"), filepath.Join(root, "imported.skel")
	input := api.Input{SkelIn: path, Sources: map[string][]byte{
		path: []byte(source),
		imported: []byte(`domain foreign.contract
pub data Value { value: string }
pub data Other { value: int }
pub data Box<TItem> { item: TItem }
pub enum Status { READY }
pub resource Record {
    check owns { input { id: string } }
    action read
    action write
}
`),
	}}
	if resolved {
		input.SkelImports = map[string]string{"foreign.contract": imported}
	}
	result, err := api.QuerySchema(input, api.SchemaQueryOption{ResolveImports: resolved})
	if err != nil {
		t.Fatal(err)
	}
	return result.Domain
}

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
	ref := before.Data()[0].Members[0].Type.List.Element
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

func TestCompareTypesAcrossResolutionStates(t *testing.T) {
	for _, test := range []struct {
		name, before, after string
		code                string
		impact              diff.ImpactLevel
	}{
		{"data", "imported.Value", "imported.Value", "", ""},
		{"enum", "imported.Status", "imported.Status", "", ""},
		{"nested generic", "map<string, list<imported.Box<imported.Value>>>", "map<string, list<imported.Box<imported.Value>>>", "", ""},
		{"different reference", "imported.Value", "imported.Other", "method.argument.type.changed", diff.ImpactBreaking},
		{"different generic argument", "imported.Box<imported.Value>", "imported.Box<imported.Other>", "method.argument.type.changed", diff.ImpactBreaking},
		{"input nullable relaxed", "imported.Value", "imported.Value?", "method.argument.type.changed", diff.ImpactCompatible},
		{"input nullable tightened", "imported.Value?", "imported.Value", "method.argument.type.changed", diff.ImpactBreaking},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, beforeResolved := range []bool{false, true} {
				for _, afterResolved := range []bool{false, true} {
					t.Run(fmt.Sprintf("resolved=%t/%t", beforeResolved, afterResolved), func(t *testing.T) {
						source := func(kind string) string {
							return "domain demo\nimport foreign.contract as imported\nservice ReadService { method read { input { value: " + kind + " } } }"
						}
						before := inspectImportedSchema(t, source(test.before), beforeResolved)
						after := inspectImportedSchema(t, strings.ReplaceAll(source(test.after), "imported", "renamed"), afterResolved)
						report, err := diff.Compare(before, after)
						if err != nil {
							t.Fatal(err)
						}
						if test.code == "" {
							if len(report.Changes) != 0 {
								t.Fatalf("resolution changed the contract: %+v", report.Changes)
							}
						} else if len(report.Changes) != 1 || report.Changes[0].Code != test.code || report.Changes[0].Impact != test.impact {
							t.Fatalf("lost real type change: %+v", report.Changes)
						}
					})
				}
			}
		})
	}
}

func TestComparePermissionsAcrossResolutionStates(t *testing.T) {
	const read = "imported.Record:read"
	const check = read + ":owns(id)"
	for _, test := range []struct {
		before, after string
		impact        diff.ImpactLevel
	}{
		{read, read, ""},
		{check, check, ""},
		{"all(" + check + ", imported.Record:write)", "all(" + check + ", imported.Record:write)", ""},
		{"any(" + check + ", imported.Record:write)", "any(" + check + ", imported.Record:write)", ""},
		{read, check, diff.ImpactBreaking},
		{check, read, diff.ImpactDangerous},
		{check, read + ":owns(otherId)", diff.ImpactDangerous},
	} {
		for _, beforeResolved := range []bool{false, true} {
			for _, afterResolved := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s->%s/resolved=%t/%t", test.before, test.after, beforeResolved, afterResolved), func(t *testing.T) {
					source := func(policy string) string {
						return `domain demo
import foreign.contract as imported
actor ClientActor { via client {} permission {} }
service ReadService {
    for ClientActor
    require imported.Record:read
    method read {
        require ` + policy + `
        input { id: string otherId: string }
    }
}`
					}
					before := inspectImportedSchema(t, source(test.before), beforeResolved)
					after := inspectImportedSchema(t, strings.ReplaceAll(source(test.after), "imported", "renamed"), afterResolved)
					report, err := diff.Compare(before, after)
					if err != nil {
						t.Fatal(err)
					}
					if test.impact == "" {
						if len(report.Changes) != 0 {
							t.Fatalf("resolution changed the permission contract: %+v", report.Changes)
						}
					} else if len(report.Changes) != 1 || report.Changes[0].Code != "method.require.changed" || report.Changes[0].Impact != test.impact {
						t.Fatalf("lost real policy change: %+v", report.Changes)
					}
					if !beforeResolved {
						term := before.Services()[0].Require.Expression
						if term.Mode != "" || term.Check.ResourceSkelName != "imported.Record" {
							t.Fatal("comparison mutated unresolved permission term")
						}
					}
				})
			}
		}
	}
}

func TestCompareActorCapabilitySourcePositions(t *testing.T) {
	for _, capability := range []string{"auth", "permission"} {
		t.Run(capability, func(t *testing.T) {
			body := "permission {}"
			if capability == "auth" {
				body = "auth { credential { token: string } info { id: string } }"
			}
			path := filepath.Join(t.TempDir(), "actor.skel")
			without := inspectSchema(t, path, "domain demo\nactor UserActor {\n    via client {}\n}\n")
			with := inspectSchema(t, path, "domain demo\nactor UserActor {\n    via client {}\n    "+body+"\n}\n")
			actorPos := schema.Position{File: path, Line: 2, Column: 7}
			sectionPos := schema.Position{File: path, Line: 4, Column: 5}
			moved := inspectSchema(t, path, "domain demo\nactor UserActor {\n    via client {}\n\n    "+body+"\n}\n")
			report, err := diff.Compare(with, moved)
			if err != nil || len(report.Changes) != 0 {
				t.Fatalf("moving a capability changed its contract: %+v, %v", report, err)
			}
			for _, added := range []bool{true, false} {
				before, after := without, with
				beforePos, afterPos := actorPos, sectionPos
				code := "actor." + capability + ".added"
				if !added {
					before, after = with, without
					beforePos, afterPos = sectionPos, actorPos
					code = "actor." + capability + ".removed"
				}
				report, err := diff.Compare(before, after)
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Changes) != 1 || report.Changes[0].Code != code {
					t.Fatalf("unexpected capability changes: %+v", report.Changes)
				}
				change := report.Changes[0]
				if change.Baseline == nil || *change.Baseline != beforePos || change.Candidate == nil || *change.Candidate != afterPos {
					t.Fatalf("wrong capability positions: %+v", change)
				}
			}
		})
	}
}

func TestCompareActorIdentifierSourcePosition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "actor.skel")
	source := `domain demo
actor UserActor {
    via client {}
    auth {
        credential { token: string }
        info { id: string }
    }
}`
	before := inspectSchema(t, path, source)
	after := inspectSchema(t, path, strings.Replace(source, "info { id:", "info { @identifier id:", 1))
	report, err := diff.Compare(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Changes) != 1 || report.Changes[0].Code != "actor.identifier.changed" {
		t.Fatalf("unexpected identifier change: %+v", report.Changes)
	}
	change := report.Changes[0]
	if change.Baseline == nil || *change.Baseline != before.Actors()[0].Auth.Pos ||
		change.Candidate == nil || *change.Candidate != after.Actors()[0].Auth.Info.Members[0].Pos {
		t.Fatalf("wrong identifier positions: %+v", change)
	}
}

func TestCompareActorAuthSensitivity(t *testing.T) {
	source := `domain demo
actor UserActor {
    via client {}
    auth {
        credential { token: string }
        info { id: string }
    }
}`
	for _, section := range []string{"credential", "info"} {
		t.Run(section, func(t *testing.T) {
			root := t.TempDir()
			beforePath, afterPath := filepath.Join(root, "before.skel"), filepath.Join(root, "after.skel")
			before := inspectSchema(t, beforePath, source)
			after := inspectSchema(t, afterPath, strings.Replace(source, section+" {", "@sensitive "+section+" {", 1))
			line := 5
			if section == "info" {
				line = 6
			}
			for _, pair := range [][2]*schema.Domain{{before, after}, {after, before}} {
				report, err := diff.Compare(pair[0], pair[1])
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Changes) != 1 || report.Changes[0].Code != "actor.auth-"+section+".sensitive.changed" || report.Changes[0].Impact != diff.ImpactDangerous || report.Changes[0].Symbol != "demo.UserActor."+section {
					t.Fatalf("unexpected sensitivity change: %+v", report.Changes)
				}
				change := report.Changes[0]
				if change.Baseline == nil || change.Candidate == nil || change.Baseline.Line != line || change.Candidate.Line != line || change.Baseline.File == change.Candidate.File {
					t.Fatalf("sensitivity change lost its section positions: %+v", change)
				}
			}
		})
	}
}
