package api_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/schema"
	"go.yorun.ai/skel/schema/diff"
)

func TestIgnoredViaAcrossGenerators(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "domain.skel")
	clean := `domain demo.access
pub actor ClientActor {
    auth { credential { token: string } info { @identifier id: string } }
    permission {}
}
pub data Item { id: int }
api service EntryApiService {
    for ClientActor
    auth required
    method get { output Item }
}
pub service EntryService { method get { output Item } }
web EntryWeb { for ClientActor auth required mount /entry }
`
	legacy := strings.Replace(clean, "pub actor ClientActor {", "pub actor ClientActor {\n via client {}\n via client {}\n via UNKNOWN {}", 1)
	legacy = strings.Replace(legacy, "permission {}", "permission {}\n via Custom {}", 1)
	legacy = strings.ReplaceAll(legacy, "for ClientActor", "for ClientActor via legacy\n for ClientActor via agent")
	parse := func(source string) api.ParseResult {
		t.Helper()
		result, err := api.Parse(api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(source)}})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	original, ignored := parse(clean), parse(legacy)
	if len(original.Diagnostics) != 0 || len(ignored.Diagnostics) != 8 {
		t.Fatalf("unexpected warnings: clean=%v legacy=%v", original.Diagnostics, ignored.Diagnostics)
	}
	for _, warning := range ignored.Diagnostics {
		if warning.Code != diagnostic.CodeViaIgnored || warning.Severity != diagnostic.SeverityWarning || warning.Range.Start.File != path || warning.Range.Start.Line < 2 || warning.Range.End.Column <= warning.Range.Start.Column {
			t.Fatalf("missing source-aware via warning: %+v", warning)
		}
	}
	if len(ignored.Domain.Actors()[0].Vias) != 0 || len(ignored.Domain.Services()[0].Audiences) != 1 || len(ignored.Domain.Webs()[0].Audiences) != 1 {
		t.Fatal("legacy vias were not discarded or audiences were not deduplicated")
	}
	report, err := diff.Compare(original.Domain, ignored.Domain)
	if err != nil {
		t.Fatal(err)
	}
	if original.Domain.Hash() != ignored.Domain.Hash() || len(report.Changes) != 0 {
		t.Fatal("ignored via changed the contract hash or compatibility report")
	}
	formatted, err := api.FormatSource([]byte(legacy))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(formatted), "via UNKNOWN {}") || len(parse(string(formatted)).Diagnostics) != 8 {
		t.Fatal("formatter lost legacy syntax or its warnings")
	}
	// Programmatic callers may still supply legacy fields. No binding may emit them.
	ignored.Domain.Actors()[0].Vias = []*schema.ActorVia{nil, {Name: "UNSUPPORTED"}}
	ignored.Domain.Services()[0].Audiences[0].Via = "UNSUPPORTED"
	ignored.Domain.Webs()[0].Audiences[0].Via = "UNSUPPORTED"
	for _, target := range []string{"go", "go-api", "go-public", "go-split", "ts", "skel"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "generated")
			generate := func(domain *schema.Domain) {
				t.Helper()
				var err error
				switch target {
				case "ts":
					err = api.GenerateTypeScript(domain, api.TypeScriptOption{ApiOnly: true, Out: out})
				case "skel":
					err = api.GenerateSkeleton(domain, api.SkeletonOption{PubOnly: true, Out: out})
				default:
					option := api.GolangOption{Out: out, CompilerVersion: "v0.0.0-dev", ApiOnly: target == "go-api", PubOnly: target == "go-public"}
					if target == "go-split" {
						option.AsModule = true
						option.Module = "example.com/access"
						option.PubOut = out + "-pub"
						option.PubModule = "example.com/accesspub"
					}
					err = api.GenerateGolang(domain, option)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			generate(original.Domain)
			before := apiOutputSnapshot(t, out)
			var beforePub map[string]string
			if target == "go-split" {
				beforePub = apiOutputSnapshot(t, out+"-pub")
			}
			generate(ignored.Domain)
			if !reflect.DeepEqual(before, apiOutputSnapshot(t, out)) || target == "go-split" && !reflect.DeepEqual(beforePub, apiOutputSnapshot(t, out+"-pub")) {
				t.Fatal("ignored via changed generated output")
			}
			for name, content := range before {
				if strings.Contains(content, "ActorVia") || strings.Contains(content, "Via:") || strings.Contains(content, "Vias:") || strings.Contains(content, " via ") {
					t.Fatalf("via metadata leaked into %s: %s", name, content)
				}
			}
			if target == "skel" {
				parsed, err := api.Parse(api.Input{SkelIn: out})
				if err != nil || len(parsed.Diagnostics) != 0 {
					t.Fatalf("generated public contract failed: %v / %v", err, parsed.Diagnostics)
				}
			}
		})
	}
}
