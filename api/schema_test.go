package api_test

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/schema"
)

func TestQuerySchemaEffectivePolicyWithOpaqueAndResolvedImports(t *testing.T) {
	root := t.TempDir()
	path, dependency := filepath.Join(root, "source.skel"), filepath.Join(root, "shared.skel")
	input := api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(`domain demo
import shared as dep
actor ClientActor { via client {} permission {} }
api service DocumentApiService {
    for ClientActor
    auth optional
    require dep.Document:read
    method get {
        require dep.Document:read:byId(id)
        input { id: string }
    }
}
`), dependency: []byte(`domain shared
pub resource Document {
    check byId { input { id: string } }
    action read {}
}
`)}}
	for _, resolved := range []bool{false, true} {
		if resolved {
			input.SkelImports = map[string]string{"shared": dependency}
		}
		result, err := api.QuerySchema(input, api.SchemaQueryOption{ResolveImports: resolved})
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.ValidateEffectivePolicy(result.Domain); err != nil {
			t.Fatal(err)
		}
		method := result.Domain.Services()[0].Methods[0]
		if method.NormalizedAuth() != schema.AuthModeInherit || method.EffectiveAuthMode != schema.AuthModeOptional {
			t.Fatalf("lost inherited policy: %+v", method)
		}
		expression := method.EffectiveRequire.Expression
		if expression.Mode != schema.PermissionRequireModeAll || len(expression.Children) != 2 {
			t.Fatalf("expected conjunction of service and method requirements: %+v", expression)
		}
		check := expression.Children[1].Children[1].Check
		if check == nil || len(check.Arguments) != 1 || check.Arguments[0].JsonPath != "id" ||
			(check.ServiceSkelName != "") != resolved || (check.Arguments[0].Type != nil) != resolved {
			t.Fatalf("unexpected check resolution (resolved=%v): %+v", resolved, check)
		}
	}
}

func TestQuerySchemaViews(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	input := api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(`domain demo
actor UserActor { via client {} }
data Hidden { value: string }
pub data Visible { value: string }
api service ReadApiService { for UserActor via client auth optional method read { output Hidden } }
`)}}
	for _, tc := range []struct {
		option          api.SchemaQueryOption
		hidden, service bool
	}{
		{api.SchemaQueryOption{}, true, true},
		{api.SchemaQueryOption{Pub: true}, false, false},
		{api.SchemaQueryOption{Api: true, ApiFilter: api.ApiFilter{Prune: true, Actors: []string{"demo.UserActor"}}}, true, true},
	} {
		result, err := api.QuerySchemaContext(t.Context(), input, tc.option)
		if err != nil {
			t.Fatal(err)
		}
		if (result.Domain.Find(schema.DeclarationTypeData, "demo.Hidden") != nil) != tc.hidden {
			t.Fatal("wrong hidden selection")
		}
		if (result.Domain.Find(schema.DeclarationTypeService, "demo.ReadApiService") != nil) != tc.service {
			t.Fatal("wrong service selection")
		}
		if len(result.Domain.Declarations()) == 0 {
			t.Fatal("empty entries")
		}
	}
	for _, option := range []api.SchemaQueryOption{{Api: true, Pub: true}, {ApiFilter: api.ApiFilter{Actors: []string{"demo.UserActor"}}}} {
		if _, err := api.QuerySchema(input, option); err == nil {
			t.Fatal("invalid selection accepted")
		}
	}
	parsed, err := api.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Domain.Find(schema.DeclarationTypeData, "demo.Visible") == nil {
		t.Fatal("parsed schema cannot be queried")
	}
}

func TestDiffSchemaFrozenSources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	baseline := api.Input{SkelIn: path, Strict: true, Sources: map[string][]byte{path: []byte("domain demo\npub service LegacyService { method ping {} }\npub data Value { value: string }\n")}}
	candidate := api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte("domain demo\npub service LegacyService { method ping {} }\npub data Value { value: int }\n")}}
	report, err := api.DiffSchemaSources(candidate, api.SchemaDiffOption{Baseline: &baseline})
	if err != nil || report.Compatible || report.Summary.Breaking == 0 {
		t.Fatalf("report=%+v, err=%v", report, err)
	}
	if report.Changes[0].Candidate == nil || report.Changes[0].Candidate.File != path {
		t.Fatalf("lost source positions: %+v", report.Changes)
	}
	if _, err := api.DiffSchemaSources(candidate, api.SchemaDiffOption{}); err == nil {
		t.Fatal("frozen Git baseline accepted")
	}
	candidate.Sources[path] = []byte("domain demo\npub data Broken {")
	if _, err := api.DiffSchemaSources(candidate, api.SchemaDiffOption{Baseline: &baseline}); !errors.Is(err, api.ErrSchemaSourceCompilation) {
		t.Fatalf("lost compilation classification: %v", err)
	}
}

func TestDiffSchemaGitBaseline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.skel")
	writeTestFile(t, path, "domain demo\npub data Value { value: string }\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "source.skel"}, {"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-qm", "baseline"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %s, %v", output, err)
		}
	}
	writeTestFile(t, path, "domain demo\npub data Value { value: int }\n")
	report, err := api.DiffSchemaSourcesContext(t.Context(), api.Input{SkelIn: path}, api.SchemaDiffOption{})
	if err != nil || report.Compatible {
		t.Fatalf("Git diff=%+v, %v", report, err)
	}
	missing := filepath.Join(t.TempDir(), "source.skel")
	writeTestFile(t, missing, "domain demo\n")
	if _, err := api.DiffSchemaSources(api.Input{SkelIn: missing}, api.SchemaDiffOption{}); !errors.Is(err, api.ErrGitHistoryUnavailable) {
		t.Fatalf("Git error lost: %v", err)
	}
}

func TestDiffSchemaErrorClassification(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeTestFile(t, path, "domain demo\n")
	input := api.Input{SkelIn: path}
	for _, baseline := range []*api.Input{nil, &input} {
		option := api.SchemaDiffOption{Baseline: baseline}
		for _, invalid := range []api.Input{{}, {SkelIn: path, SkelImports: map[string]string{"other": path}}} {
			_, err := api.DiffSchemaSources(invalid, option)
			if err == nil || errors.Is(err, api.ErrSchemaSourceCompilation) {
				t.Fatalf("invalid options classified as source compilation: %v", err)
			}
		}
		canceled, cancel := context.WithCancel(t.Context())
		cancel()
		expired, stop := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
		defer stop()
		for _, ctx := range []context.Context{canceled, expired} {
			_, err := api.DiffSchemaSourcesContext(ctx, input, option)
			if !errors.Is(err, ctx.Err()) || errors.Is(err, api.ErrSchemaSourceCompilation) {
				t.Fatalf("cancellation classified as source compilation: %v", err)
			}
		}
	}
	writeTestFile(t, path, "domain demo\npub data Broken {")
	for _, baseline := range []*api.Input{nil, &input} {
		_, err := api.DiffSchemaSources(input, api.SchemaDiffOption{Baseline: baseline})
		if !errors.Is(err, api.ErrSchemaSourceCompilation) {
			t.Fatalf("source error lost classification: %v", err)
		}
	}
}
