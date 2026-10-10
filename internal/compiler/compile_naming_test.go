package compiler

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeclarationVersionSuffixes(t *testing.T) {
	for _, declaration := range []struct{ prefix, suffix, body string }{
		{"pub service", "Service", "method ping {}"},
		{"ext service", "Service", "method ping {}"},
		{"api service", "ApiService", "for ClientActor via client auth required method ping {}"},
		{"pub event", "Event", "payload { id: int }"},
		{"ext event", "Event", "payload { id: int }"},
		{"config", "Config", "id: int"},
		{"pub actor", "Actor", "via client {}"},
		{"task", "Task", "trigger run {}"},
		{"web", "Web", "for ClientActor via client auth required"},
	} {
		for _, version := range []string{"", "V1", "V2", "V10", "V999999999999999999999", "V0", "V01", "V", "v2", "V2Beta", "V2V3"} {
			t.Run(declaration.prefix+"/"+version, func(t *testing.T) {
				name := "Example" + declaration.suffix + version
				qualifier := ""
				if declaration.suffix == "Config" {
					qualifier = " eternal"
				}
				path := filepath.Join(t.TempDir(), "domain.skel")
				writeFile(t, path, fmt.Sprintf("domain demo\nactor ClientActor { via client {} }\n%s %s%s { %s }\n", declaration.prefix, name, qualifier, declaration.body))
				valid := version == "" || version == "V1" || version == "V2" || version == "V10" || version == "V999999999999999999999"
				result, err := Check(Option{SkelIn: path})
				if err != nil {
					t.Fatal(err)
				}
				if result.Diagnostics.HasErrors() == valid {
					t.Fatalf("valid=%v, diagnostics: %v", valid, result.Diagnostics)
				}
				for _, diagnostic := range result.Diagnostics {
					if diagnostic.Range.Start.File != path || diagnostic.Range.Start.Line != 3 {
						t.Fatalf("unexpected diagnostic location: %+v", diagnostic)
					}
				}
				if _, err := Compile(Option{SkelIn: path}); (err == nil) != valid {
					t.Fatalf("valid=%v, compile error: %v", valid, err)
				}
			})
		}
	}
}

func TestEventAndConfigNamesRequireBody(t *testing.T) {
	for _, declaration := range []struct{ prefix, suffix, qualifier, body string }{
		{"event", "Event", "", "payload { id: int }"},
		{"pub event", "Event", "", "payload { id: int }"},
		{"ext event", "Event", "", "payload { id: int }"},
		{"config", "Config", " eternal", "id: int"},
		{"config", "Config", " instant", "id: int"},
		{"pub config", "Config", " eternal", "id: int"},
		{"pub config", "Config", " instant", "id: int"},
	} {
		for _, version := range []string{"", "V1", "V2", "V10"} {
			t.Run(declaration.prefix+declaration.qualifier+"/"+version, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "domain.skel")
				writeFile(t, path, fmt.Sprintf("domain demo\n%s %s%s%s { %s }\n", declaration.prefix, declaration.suffix, version, declaration.qualifier, declaration.body))
				result, err := Check(Option{SkelIn: path})
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Diagnostics) != 1 || !result.Diagnostics.HasErrors() || !strings.Contains(result.Diagnostics[0].Message, "missing body") {
					t.Fatalf("expected missing name body diagnostic, got %+v", result.Diagnostics)
				}
				if pos := result.Diagnostics[0].Range.Start; pos.File != path || pos.Line != 2 || pos.Column != len(declaration.prefix)+2 {
					t.Fatalf("unexpected diagnostic location: %+v", pos)
				}
				if _, err := Compile(Option{SkelIn: path}); err == nil {
					t.Fatal("accepted declaration without a name body")
				}
			})
		}
	}
}

func TestVersionSuffixPreservesNamingRules(t *testing.T) {
	for _, declaration := range []string{
		"pub service ServiceV2 { method ping {} }",
		"api service ApiServiceV2 { for ClientActor auth required method ping {} }",
		"api service OrderServiceV2 { for ClientActor auth required method ping {} }",
		"api service OrderAPIServiceV2 { for ClientActor auth required method ping {} }",
		"actor ActorV2 { via client {} }",
		"task TaskV2 { trigger run {} }",
		"web WebV2 { for ClientActor auth required }",
	} {
		t.Run(declaration, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "domain.skel")
			writeFile(t, path, "domain demo\nactor ClientActor { via client {} }\n"+declaration)
			if _, err := Compile(Option{SkelIn: path}); err == nil {
				t.Fatal("accepted declaration without the required name body or kind suffix")
			}
		})
	}
	// Existing ordinary identifiers and version-before-kind names remain valid.
	parseDomain(t, map[string]string{"domain.skel": "domain demo\n", "names.skel": `domain demo
pub service OrderV2Service { method ping {} }
data OrderServiceV2 {}
enum StateEventV2 { READY }
`})
}

func TestVersionedNameSuggestionPreservesVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, "domain demo\npub service orderServiceV10 { method ping {} }\n")
	result, err := Check(Option{SkelIn: path})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Suggestion == nil || result.Diagnostics[0].Suggestion.Replacement != "OrderServiceV10" {
		t.Fatalf("unexpected naming suggestion: %+v", result.Diagnostics)
	}
}
