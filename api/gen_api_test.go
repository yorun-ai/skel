package api_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/testutil"
)

func TestApiGoClientCrossDomainAndInvocation(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	shared := filepath.Join(root, "shared.skel")
	order := filepath.Join(root, "order.skel")
	writeTestFile(t, shared, `domain common.shared
pub data Page<TItem> { items: list<TItem> }
data Detail { label: string }
pub data External { detail: Detail }
`)
	writeTestFile(t, order, `domain shop.order
import common.shared
actor ClientActor { via client {} }
data Unused { secret: string }
api service OrderApiService {
    for ClientActor via client
    method get {
        input { payload: common.shared.Page<common.shared.Page<binary>> }
        output common.shared.External
    }
    method echo {
        input {
        value: binary
        ctx: string
        options: int
        result: string
        client: string
        ret: string
        err: string
    }
        output binary
    }
    method ping {}
}
api service HealthApiService { for ClientActor via client method ping {} }
pub service BackendService { method ping {} }
`)
	sharedOut, out := filepath.Join(root, "sharedapi"), filepath.Join(root, "orderapi")
	if _, err := api.CompileGolang(api.Input{SkelIn: shared}, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, AsModule: true, Out: sharedOut, ModulePrefix: "example.com/gen"}); err != nil {
		t.Fatal(err)
	}
	parsed, err := api.Parse(api.Input{SkelIn: order, SkelImports: map[string]string{"common.shared": shared}})
	if err != nil {
		t.Fatal(err)
	}
	for _, service := range parsed.Domain.Services() {
		if service.Name != "OrderApiService" {
			continue
		}
		for _, method := range service.Methods {
			if method.Name != "get" {
				continue
			}
			method.Description = "Get external data"
			method.Deprecated = true
			method.DeprecatedReason = "Use fetch"
			method.Arguments[0].Description = "Nested payload"
			method.Arguments[0].Example = "payload example"
			method.OutputDescription = "External result"
			method.OutputExample = "result example"
		}
	}
	if err := api.GenerateGolang(parsed.Domain, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, AsModule: true, Out: out, ModulePrefix: "example.com/gen"}); err != nil {
		t.Fatal(err)
	}
	tsOut := filepath.Join(root, "typescript")
	if _, err := api.CompileTypeScript(api.Input{SkelIn: order, SkelImports: map[string]string{"common.shared": shared}}, api.TypeScriptOption{ApiOnly: true, Out: tsOut, Imports: map[string]string{"common.shared": "@demo/sharedapi"}}); err != nil {
		t.Fatal(err)
	}
	spec, err := os.ReadFile(filepath.Join(tsOut, "spec.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(spec), "createPageWireSchema(createPageWireSchema({ kind: 'binary' }))") {
		t.Fatalf("nested generic binary schema is missing: %s", spec)
	}
	service, err := os.ReadFile(filepath.Join(out, "service.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(service), "func init()") != 1 || !strings.Contains(string(service), "vrpc.Register(_HealthApiServiceSpec)") || !strings.Contains(string(service), "vrpc.Register(_OrderApiServiceSpec)") {
		t.Fatalf("expected shared init and standalone service specs: %s", service)
	}
	if !strings.Contains(string(service), `_OrderApiServiceGetMethod = vrpc.MustGetMethodInfo("shop.order.OrderApiService", "get")`) {
		t.Fatalf("missing required method lookup: %s", service)
	}
	mod, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "go.yorun.ai/vrpc v0.13.0") {
		t.Fatalf("unexpected API runtime dependency: %s", mod)
	}

	if strings.Contains(string(service), "BackendService") || strings.Contains(string(service), "ResponseMetadata") || strings.Contains(string(service), "go.yorun.ai/vine") {
		t.Fatalf("unexpected API client: %s", service)
	}
	data, err := os.ReadFile(filepath.Join(sharedOut, "data.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "type Detail struct") {
		t.Fatalf("missing implicit dependency: %s", data)
	}
	if strings.Contains(string(data), "OrderApiServiceGetArguments") || strings.Contains(string(service), "type OrderApiServiceGetArguments") {
		t.Fatal("API arguments must not be public data types")
	}
	if !strings.Contains(string(service), "type _OrderApiServiceGetArguments struct") || !strings.Contains(string(service), "payload sharedapi.Page[sharedapi.Page[skeltype.Binary]]") {
		t.Fatalf("missing positional API arguments: %s", service)
	}
	for _, fragment := range []string{"Get external data", "@param payload - Nested payload", "payload example", "@returns", "External result", "result example", "Deprecated: Use fetch."} {
		if !strings.Contains(string(service), fragment) {
			t.Errorf("missing API method documentation %q", fragment)
		}
	}
	writeTestFile(t, filepath.Join(out, "invoke_test.go"), apiInvocationTest)
	testutil.Go(t, out, "mod", "edit", "-replace=example.com/gen/common/sharedapi="+sharedOut)
	testutil.UseLocalSkel(t, out)
	deps := testutil.Go(t, out, "list", "-mod=mod", "-deps", "./...")
	if strings.Contains(deps, "go.yorun.ai/vine") {
		t.Fatal("API module depends on Vine")
	}
	testutil.Go(t, out, "test", "-mod=mod", "./...")
}

const apiInvocationTest = `package orderapi
import (
    "context"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/fxamacker/cbor/v2"
    "go.yorun.ai/vrpc"
    "go.yorun.ai/skel/types"
    shared "example.com/gen/common/sharedapi"
)
func TestInvoke(t *testing.T) {
    identity := vrpc.Identity{Name: "demo.server", Version: "1.0.0", InstanceID: "550e8400-e29b-41d4-a716-446655440000"}
    serverHeader, err := vrpc.EncodeIdentity(identity)
    if err != nil { t.Fatal(err) }
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("vrpc-status", "OK")
        w.Header().Set("vrpc-server", serverHeader)
        w.Header().Set("content-type", "application/vrpc+json")
        switch r.URL.Path {
        case "/invoke/shop.order.OrderApiService/get":
            if r.Header.Get("content-type") != "application/vrpc+cbor" { t.Error("generic binary argument did not select CBOR") }
            io.WriteString(w, "{\"result\":{\"detail\":{\"label\":\"ok\"}}}")
        case "/invoke/shop.order.OrderApiService/echo":
            bodyBytes, readErr := io.ReadAll(r.Body)
            if readErr != nil { t.Fatal(readErr) }
            var request struct {
                Params struct {
                    Value []byte ` + "`json:\"value\"`" + `
                    Context string ` + "`json:\"ctx\"`" + `
                    Options int32 ` + "`json:\"options\"`" + `
                    Result string ` + "`json:\"result\"`" + `
                } ` + "`json:\"params\"`" + `
            }
            if err := cbor.Unmarshal(bodyBytes, &request); err != nil { t.Fatal(err) }
            if request.Params.Context != "context" || request.Params.Options != 7 || request.Params.Result != "result" || len(request.Params.Value) != 2 {
                t.Fatalf("unexpected arguments: %+v", request.Params)
            }
            w.Header().Set("content-type", "application/vrpc+cbor")
            body, err := cbor.Marshal(map[string]any{"result": []byte{0,255}})
            if err != nil { t.Error(err) }
            w.Write(body)
        case "/invoke/shop.order.OrderApiService/ping":
            io.WriteString(w, "{\"result\":null}")
        default: t.Errorf("unexpected path: %s", r.URL.Path)
        }
    }))
    defer server.Close()
    raw, err := vrpc.NewClient(vrpc.Option{Endpoint: server.URL+"/invoke", Identity: identity})
    if err != nil { t.Fatal(err) }
    var client OrderApiServiceClient = NewOrderApiServiceClient(raw)
    got, err := client.Get(context.Background(), shared.Page[shared.Page[types.Binary]]{Items: []shared.Page[types.Binary]{{Items: []types.Binary{{0,255}}}}})
    if err != nil || got.Detail.Label != "ok" { t.Fatalf("get: %+v %v", got, err) }
    blob, err := client.Echo(context.Background(), types.Binary{0,255}, "context", 7, "result", "client", "ret", "err")
    if err != nil || len(blob) != 2 || blob[1] != 255 { t.Fatalf("echo: %v %v", blob, err) }
    if err := client.Ping(context.Background()); err != nil { t.Fatal(err) }
    fake := &fakeOrderClient{}
    client = fake
    if err := client.Ping(context.Background()); err != nil || !fake.called {
        t.Fatalf("replacement client was not called: %v", err)
    }
}

type fakeOrderClient struct {
    OrderApiServiceClient
    called bool
}

func (client *fakeOrderClient) Ping(ctx context.Context, options ...vrpc.InvokeOption) error {
    client.called = true
    return nil
}
`

func TestApiOutputBoundaryAndUnusedBackendImport(t *testing.T) {
	root := t.TempDir()
	backend := filepath.Join(root, "backend.skel")
	entry := filepath.Join(root, "order.skel")
	writeTestFile(t, backend, "domain demo.backend\npub data Internal { id: string }\n")
	writeTestFile(t, entry, `domain demo.order
import demo.backend as backend
data Item { id: string }
data Unused { hidden: string }
pub enum State { READY }
actor TestActor { via client {} }
api service OrderApiService { for TestActor via client method get { output Item } }
service LegacyService { method get { noauth output Item } }
pub service BackendService { method get { output backend.Internal } }
service HiddenService { method ping {} }
`)
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"demo.backend": backend}}
	goOut, tsOut := filepath.Join(root, "orderapi"), filepath.Join(root, "ts")
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, AsModule: true, Module: "example.com/orderapi", Out: goOut}); err != nil {
		t.Fatal(err)
	}
	if _, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, Out: tsOut}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(goOut, "service.go"), filepath.Join(tsOut, "service.ts")} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		code := string(contents)
		if !strings.Contains(code, "OrderApiService") || !strings.Contains(code, "LegacyService") || strings.Contains(code, "BackendService") || strings.Contains(code, "HiddenService") {
			t.Fatalf("wrong service selection in %s: %s", path, code)
		}
	}
	for _, path := range []string{filepath.Join(goOut, "data.go"), filepath.Join(tsOut, "data.ts")} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(contents), "Item") || strings.Contains(string(contents), "Unused") || strings.Contains(string(contents), "Internal") {
			t.Fatalf("wrong data selection in %s: %s", path, contents)
		}
	}
	if _, err := api.CompileTypeScript(input, api.TypeScriptOption{Out: tsOut}); err == nil {
		t.Fatal("accepted TypeScript generation without api")
	}
	if _, err := api.CompileGolang(input, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, PubOnly: true, Out: goOut}); err == nil {
		t.Fatal("accepted both Go modes")
	}
}

func TestCrossDomainImportAliasCollisions(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	imports := map[string]string{}
	goImports := map[string]string{}
	tsImports := map[string]string{}
	outputs := map[string]string{}
	for _, prefix := range []string{"first", "second"} {
		domain := prefix + ".user"
		input := filepath.Join(root, prefix+".skel")
		writeTestFile(t, input, "domain "+domain+"\npub data Value { id: string }\n")
		imports[domain] = input
		goImports[domain] = "example.com/" + prefix + "/userapi"
		tsImports[domain] = "./" + prefix
		outputs[domain] = filepath.Join(root, prefix)
		if _, err := api.CompileGolang(api.Input{SkelIn: input}, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, AsModule: true, Module: goImports[domain], Out: outputs[domain]}); err != nil {
			t.Fatal(err)
		}
	}
	var firstGo, firstTs string
	for _, reversed := range []bool{false, true} {
		declarations := "import first.user\nimport second.user\n"
		if reversed {
			declarations = "import second.user\nimport first.user\n"
		}
		input := filepath.Join(t.TempDir(), "app.skel")
		writeTestFile(t, input, "domain demo.app\n"+declarations+"data Pair { first: first.user.Value second: second.user.Value }\nactor TestActor { via client {} }\napi service AppApiService { for TestActor via client method get { output Pair } }\n")
		out := filepath.Join(t.TempDir(), "appapi")
		if _, err := api.CompileGolang(api.Input{SkelIn: input, SkelImports: imports}, api.GolangOption{CompilerVersion: "v0.0.0-dev", ApiOnly: true, AsModule: true, Module: "example.com/appapi", Out: out, Imports: goImports}); err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(filepath.Join(out, "data.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{`firstuser "example.com/first/userapi"`, `seconduser "example.com/second/userapi"`, "firstuser.Value", "seconduser.Value"} {
			if !strings.Contains(string(content), fragment) {
				t.Fatalf("missing %s:\n%s", fragment, content)
			}
		}
		if reversed && string(content) != firstGo {
			t.Fatal("Go imports depend on declaration order")
		}
		firstGo = string(content)
		for domain, path := range outputs {
			testutil.Go(t, out, "mod", "edit", "-replace="+goImports[domain]+"="+path)
		}
		testutil.UseLocalSkel(t, out)
		testutil.Go(t, out, "test", "-mod=mod", "./...")
		tsOut := t.TempDir()
		if _, err := api.CompileTypeScript(api.Input{SkelIn: input, SkelImports: imports}, api.TypeScriptOption{ApiOnly: true, Out: tsOut, Imports: tsImports}); err != nil {
			t.Fatal(err)
		}
		ts, err := os.ReadFile(filepath.Join(tsOut, "data.ts"))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{"import type * as firstUser from './first'", "import type * as secondUser from './second'", "firstUser.Value", "secondUser.Value"} {
			if !strings.Contains(string(ts), fragment) {
				t.Fatalf("missing %s:\n%s", fragment, ts)
			}
		}
		if reversed && string(ts) != firstTs {
			t.Fatal("TypeScript imports depend on declaration order")
		}
		firstTs = string(ts)
	}
}

func TestApiActorFilterAcrossTargets(t *testing.T) {
	root := t.TempDir()
	actors, backend, entry := filepath.Join(root, "actors.skel"), filepath.Join(root, "backend.skel"), filepath.Join(root, "entry.skel")
	writeTestFile(t, actors, `domain identity.user
pub actor UserActor { via client {} via agent {} }
`)
	writeTestFile(t, backend, `domain private.backend
pub data Secret { value: string }
`)
	writeTestFile(t, entry, `domain demo.order
import identity.user as identity
import private.backend as backend
actor AdminActor { via client {} }
actor IdleActor { via client {} }
pub data PublicValue { value: string }
data UserValue { value: string }
api service UserApiService {
 for identity.UserActor via agent
 noauth
 method get { output UserValue }
}
api service SharedApiService {
 for identity.UserActor via client
 for AdminActor via client
 method ping {}
}
api service AdminApiService {
 for AdminActor via client
 method get { output backend.Secret }
}
`)
	input := api.Input{SkelIn: entry, SkelImports: map[string]string{"identity.user": actors, "private.backend": backend}}
	for _, target := range []string{"go", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			generate := func(names []string) error {
				selection := api.ApiFilter{Actors: names}
				if target == "go" {
					_, err := api.CompileGolang(input, api.GolangOption{ApiOnly: true, ApiFilter: selection, Out: out, Imports: map[string]string{"private.backend": "example.com/backendapi"}})
					return err
				}
				_, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, ApiFilter: selection, Out: out, Imports: map[string]string{"private.backend": "@demo/backendapi"}})
				return err
			}
			ext := "." + target
			if target == "ts" {
				ext = ".ts"
			}
			servicePath, dataPath := filepath.Join(out, "service"+ext), filepath.Join(out, "data"+ext)
			for _, names := range [][]string{nil, {"identity.user.UserActor"}, {"demo.order.AdminActor", "identity.user.UserActor"}, {"demo.order.IdleActor"}} {
				if err := generate(names); err != nil {
					t.Fatal(err)
				}
				service, err := os.ReadFile(servicePath)
				if len(names) == 1 && names[0] == "demo.order.IdleActor" {
					if target == "ts" {
						if err != nil || strings.Contains(string(service), "ApiService") {
							t.Fatalf("stale TS service: %s, %v", service, err)
						}
					} else if !os.IsNotExist(err) {
						t.Fatalf("stale service output: %s, %v", service, err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					for _, name := range []string{"UserApiService", "SharedApiService"} {
						if !strings.Contains(string(service), name) {
							t.Fatalf("missing %s: %s", name, service)
						}
					}
					wantAdmin := len(names) != 1
					if strings.Contains(string(service), "AdminApiService") != wantAdmin {
						t.Fatalf("wrong selection: %s", service)
					}
				}
				data, err := os.ReadFile(dataPath)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "PublicValue") || strings.Contains(string(data), "UserValue") != (len(names) != 1 || names[0] != "demo.order.IdleActor") {
					t.Fatalf("wrong data closure: %s", data)
				}
			}
			before := apiOutputSnapshot(t, out)
			for _, name := range []string{"UserActor", "identity.UserActor", "identity.user.MissingActor", ""} {
				if err := generate([]string{name}); err == nil {
					t.Fatalf("accepted invalid actor %q", name)
				}
				after := apiOutputSnapshot(t, out)
				if len(before) != len(after) {
					t.Fatal("failed generation changed output")
				}
				for path, content := range before {
					if after[path] != content {
						t.Fatalf("changed %s", path)
					}
				}
			}
			// The unselected admin service must not require an API import mapping.
			if target == "go" {
				_, err := api.CompileGolang(input, api.GolangOption{ApiOnly: true, ApiFilter: api.ApiFilter{Actors: []string{"identity.user.UserActor"}}, Out: out})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				_, err := api.CompileTypeScript(input, api.TypeScriptOption{ApiOnly: true, ApiFilter: api.ApiFilter{Actors: []string{"identity.user.UserActor"}}, Out: out})
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestExtensionServicesAreExcludedFromApiClients(t *testing.T) {
	source := filepath.Join(t.TempDir(), "service.skel")
	// Legacy admission rules must not turn an extension into a client API.
	writeTestFile(t, source, `domain demo.storage
actor ClientActor { via client {} }
ext service StorageService { noauth method get { output string } }
api service HealthApiService { for ClientActor via client noauth method ping {} }
`)
	for _, target := range []string{"go", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			var err error
			if target == "go" {
				_, err = api.CompileGolang(api.Input{SkelIn: source}, api.GolangOption{ApiOnly: true, Out: out})
			} else {
				_, err = api.CompileTypeScript(api.Input{SkelIn: source}, api.TypeScriptOption{ApiOnly: true, Out: out})
			}
			if err != nil {
				t.Fatal(err)
			}
			contents, err := os.ReadFile(filepath.Join(out, "service."+target))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(contents), "StorageService") || !strings.Contains(string(contents), "HealthApiService") {
				t.Fatalf("incorrect API boundary: %s", contents)
			}
		})
	}
}
