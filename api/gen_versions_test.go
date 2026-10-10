package api_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/testutil"
	"go.yorun.ai/skel/schema"
)

func TestVersionedDeclarationsAcrossGenerators(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "domain.skel")
	versions := []string{"", "V1", "V2", "V10"}
	source := "// Versions coexist in one domain.\ndomain demo.versions\npub data Item { id: int content: binary }\n"
	for _, version := range versions {
		source += fmt.Sprintf(`
pub actor ClientActor%[1]s {
    via client {}
    auth {
        credential { token: string }
        info { @identifier id: int }
    }
    permission {}
}
pub service OrderService%[1]s { method get { input { id: int } output Item } }
ext service StorageService%[1]s { method store { input { item: Item } } }
api service OrderApiService%[1]s {
    for ClientActor%[1]s via client
    auth required
    method get { input { id: int } output Item }
}
pub event OrderPlacedEvent%[1]s { payload { item: Item } }
ext event AuditRecordedEvent%[1]s { payload { item: Item } }
config CheckoutConfig%[1]s eternal { limit: int }
task RebuildTask%[1]s { trigger run { input { item: Item } } }
web ConsoleWeb%[1]s { for ClientActor%[1]s via client auth required }
`, version)
	}
	formatted, err := api.FormatSource([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	again, err := api.FormatSource(formatted)
	if err != nil || string(again) != string(formatted) {
		t.Fatalf("versioned source formatting is not stable: %v", err)
	}
	if !strings.Contains(string(formatted), "// Versions coexist in one domain.") {
		t.Fatal("formatting lost the source comment")
	}
	writeTestFile(t, path, string(formatted))
	input := api.Input{SkelIn: path}
	parsed, err := api.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range versions {
		service := parsed.Domain.Find(schema.DeclarationTypeService, "demo.versions.OrderApiService"+version)
		if service == nil {
			t.Fatalf("missing API service version %q", version)
		}
	}

	for _, mode := range []string{"backend", "public", "split", "api"} {
		t.Run("go/"+mode, func(t *testing.T) {
			out := filepath.Join(root, mode)
			option := api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, AsModule: true, Module: "example.com/versions", PubOnly: mode == "public", ApiOnly: mode == "api"}
			if mode == "split" {
				option.PubOut = filepath.Join(root, "splitpub")
				option.PubModule = "example.com/versionspub"
			}
			if err := api.GenerateGolang(parsed.Domain, option); err != nil {
				t.Fatal(err)
			}
			before := apiOutputSnapshot(t, out)
			if err := api.GenerateGolang(parsed.Domain, option); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, apiOutputSnapshot(t, out)) {
				t.Fatal("regeneration changed the output")
			}
			outputs := []string{out}
			if mode == "split" {
				outputs = append(outputs, option.PubOut)
			}
			for _, output := range outputs {
				var generated strings.Builder
				for name, content := range apiOutputSnapshot(t, output) {
					if strings.HasSuffix(name, ".go") {
						generated.WriteString(content)
					}
				}
				service := generated.String()
				for _, version := range versions {
					name := "OrderService" + version
					if mode == "api" {
						name = "OrderApiService" + version
					}
					if !strings.Contains(service, name+"Client") || !strings.Contains(service, `"demo.versions.`+name+`"`) {
						t.Fatalf("missing service identity %s in %s", name, output)
					}
					if mode == "api" {
						continue
					}
					event := readTestFile(t, filepath.Join(output, "event.go"))
					placedRole, placedMethod, auditRole, auditMethod := "Listener", "On", "Emitter", "Emit"
					if mode == "split" && output == out {
						placedRole, placedMethod, auditRole, auditMethod = "Emitter", "Emit", "Listener", "On"
					}
					for _, fragment := range []string{"OrderPlacedEvent" + version + placedRole, placedMethod + "OrderPlaced" + version + "(", "AuditRecordedEvent" + version + auditRole, auditMethod + "AuditRecorded" + version + "(", `"demo.versions.OrderPlacedEvent` + version + `"`} {
						if !strings.Contains(event, fragment) {
							t.Fatalf("missing event %q in %s", fragment, output)
						}
					}
					if version != "" && (strings.Contains(event, "OnOrderPlacedEvent"+version) || strings.Contains(event, "EmitAuditRecordedEvent"+version)) {
						t.Fatalf("kind suffix leaked into event method names in %s", output)
					}
					for _, fragment := range []string{"ClientActor" + version + "Credential", "ClientActor" + version + "Info", "ClientActor" + version + "AuthService", "ClientActor" + version + "PermissionService", `"demo.versions.ClientActor` + version + `"`} {
						if !strings.Contains(service, fragment) {
							t.Fatalf("missing actor %q in %s", fragment, output)
						}
					}
					if mode == "public" || output == option.PubOut {
						continue
					}
					for file, name := range map[string]string{"config.go": "CheckoutConfig", "task.go": "RebuildTask", "web.go": "ConsoleWeb"} {
						content := readTestFile(t, filepath.Join(output, file))
						if !strings.Contains(content, name+version) || !strings.Contains(content, `"demo.versions.`+name+version+`"`) {
							t.Fatalf("missing version %q in %s", version, file)
						}
					}
				}
			}
			// Backend compilation is verified in the consuming runtime; keep CI
			// independent of Vine, as with the other generator integration tests.
			if mode == "api" {
				t.Run("compile", func(t *testing.T) {
					testutil.RequireToolchain(t)
					testutil.UseLocalSkel(t, out)
					testutil.Go(t, out, "test", "-mod=mod", "./...")
				})
			}
		})
	}

	t.Run("typescript", func(t *testing.T) {
		out := filepath.Join(root, "ts")
		if err := api.GenerateTypeScript(parsed.Domain, api.TypeScriptOption{Out: out, ApiOnly: true}); err != nil {
			t.Fatal(err)
		}
		service, spec := readTestFile(t, filepath.Join(out, "service.ts")), readTestFile(t, filepath.Join(out, "spec.ts"))
		for _, version := range versions {
			if !strings.Contains(service, "createOrderApiService"+version+"(") || !strings.Contains(spec, "serviceName: 'demo.versions.OrderApiService"+version+"'") {
				t.Fatalf("missing TypeScript API version %q", version)
			}
		}
	})

	t.Run("public-skel", func(t *testing.T) {
		out := filepath.Join(root, "skel")
		if err := api.GenerateSkeleton(parsed.Domain, api.SkeletonOption{Out: out, PubOnly: true}); err != nil {
			t.Fatal(err)
		}
		roundTrip, err := api.Parse(api.Input{SkelIn: out})
		if err != nil {
			t.Fatal(err)
		}
		if len(roundTrip.Domain.Services()) != 8 || len(roundTrip.Domain.Events()) != 8 || len(roundTrip.Domain.Actors()) != 4 {
			t.Fatal("public contract lost a declaration version")
		}
		consumer := filepath.Join(root, "consumer.skel")
		writeTestFile(t, consumer, `domain demo.consumer
import demo.versions as orders
api service ConsumerApiServiceV10 {
    for orders.ClientActorV10 via client
    auth required
    method get { output orders.Item }
}
`)
		if _, err := api.Parse(api.Input{SkelIn: consumer, SkelImports: map[string]string{"demo.versions": out}}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestVersionedActorFilterAcrossGenerators(t *testing.T) {
	root := t.TempDir()
	actors, entry := filepath.Join(root, "actors.skel"), filepath.Join(root, "entry.skel")
	writeTestFile(t, actors, `domain identity.user
pub actor ClientActorV1 { via client {} }
pub actor ClientActorV10 { via client {} }
`)
	source := "domain demo.order\nimport identity.user as identity\n"
	for _, version := range []string{"V1", "V10"} {
		source += fmt.Sprintf(`
actor ClientActor%[1]s { via client {} }
data LocalResult%[1]s { id: int }
data ImportedResult%[1]s { id: int }
api service LocalApiService%[1]s {
    for ClientActor%[1]s via client auth required
    method get { output LocalResult%[1]s }
}
api service ImportedApiService%[1]s {
    for identity.ClientActor%[1]s via client auth required
    method get { output ImportedResult%[1]s }
}
`, version)
	}
	writeTestFile(t, entry, source)
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"identity.user": actors}}
	for _, target := range []string{"go", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			for _, test := range []struct {
				name   string
				actors []string
				want   []string
			}{
				{"all", nil, []string{"LocalV1", "LocalV10", "ImportedV1", "ImportedV10"}},
				{"local-v1", []string{"demo.order.ClientActorV1"}, []string{"LocalV1"}},
				{"local-v10", []string{"demo.order.ClientActorV10"}, []string{"LocalV10"}},
				{"imported-v1", []string{"identity.user.ClientActorV1"}, []string{"ImportedV1"}},
				{"imported-v10", []string{"identity.user.ClientActorV10"}, []string{"ImportedV10"}},
				{"union", []string{"demo.order.ClientActorV1", "identity.user.ClientActorV10"}, []string{"LocalV1", "ImportedV10"}},
			} {
				t.Run(test.name, func(t *testing.T) {
					selection := api.ApiFilter{Actors: test.actors}
					var err error
					if target == "go" {
						_, err = api.CompileGolang(input, api.GolangOption{ApiOnly: true, ApiFilter: selection, Out: out})
					} else {
						_, err = api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, ApiFilter: selection, Out: out})
					}
					if err != nil {
						t.Fatal(err)
					}
					service := readTestFile(t, filepath.Join(out, "service."+target))
					data := readTestFile(t, filepath.Join(out, "data."+target))
					spec := service
					if target == "ts" {
						spec = readTestFile(t, filepath.Join(out, "spec.ts"))
					}
					for _, scope := range []string{"Local", "Imported"} {
						for _, version := range []string{"V1", "V10"} {
							want := slices.Contains(test.want, scope+version)
							name := scope + "ApiService" + version
							factory, identity := "func New"+name+"Client(", `"demo.order.`+name+`"`
							if target == "ts" {
								factory, identity = "function create"+name+"(", "'demo.order."+name+"'"
							}
							// Include delimiters so V1 assertions cannot match V10.
							if strings.Contains(service, factory) != want || strings.Contains(spec, identity) != want {
								t.Errorf("service %s: want selected=%v\n%s\n%s", name, want, service, spec)
							}
							if strings.Contains(data, "type "+scope+"Result"+version+" ") != want {
								t.Errorf("wrong data closure for %s%s: %s", scope, version, data)
							}
						}
					}
				})
			}
		})
	}
}
