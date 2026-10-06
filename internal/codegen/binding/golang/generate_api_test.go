package golang_test

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
	writeFileForTest(t, shared, `domain common.shared
pub data Page<TItem> { items: list<TItem> }
data Detail { label: string }
pub data External { detail: Detail }
`)
	writeFileForTest(t, order, `domain shop.order
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
	if !strings.Contains(string(service), "type _OrderApiServiceGetArguments struct") || !strings.Contains(string(service), "payload sharedapi.Page[sharedapi.Page[skel.Binary]]") {
		t.Fatalf("missing positional API arguments: %s", service)
	}
	for _, fragment := range []string{"Get external data", "@param payload - Nested payload", "payload example", "@returns", "External result", "result example", "Deprecated: Use fetch."} {
		if !strings.Contains(string(service), fragment) {
			t.Errorf("missing API method documentation %q", fragment)
		}
	}
	writeFileForTest(t, filepath.Join(out, "invoke_test.go"), apiInvocationTest)
	testutil.Go(t, out, "mod", "edit", "-replace=example.com/gen/common/sharedapi="+sharedOut)
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
    "go.yorun.ai/vrpc/skel"
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
    got, err := client.Get(context.Background(), shared.Page[shared.Page[skel.Binary]]{Items: []shared.Page[skel.Binary]{{Items: []skel.Binary{{0,255}}}}})
    if err != nil || got.Detail.Label != "ok" { t.Fatalf("get: %+v %v", got, err) }
    blob, err := client.Echo(context.Background(), skel.Binary{0,255}, "context", 7, "result", "client", "ret", "err")
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
	writeFileForTest(t, backend, "domain demo.backend\npub data Internal { id: string }\n")
	writeFileForTest(t, entry, `domain demo.order
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

func TestApiBackendSchemaAndClientBoundary(t *testing.T) {
	root := t.TempDir()
	entry := filepath.Join(root, "order.skel")
	writeFileForTest(t, entry, `domain demo.order
actor TestActor { via client {} }
api service OrderApiService { for TestActor via client method ping {} }
pub service BackendService { method ping {} }
`)
	input := api.Input{SkelIn: entry}
	out := filepath.Join(root, "server")
	option := api.GolangOption{CompilerVersion: "v0.0.0-dev", AsModule: true, Module: "example.com/orderserver", Out: out}
	if _, err := api.CompileGolang(input, option); err != nil {
		t.Fatal(err)
	}
	schema, err := os.ReadFile(filepath.Join(out, "schema.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(schema), "Api:") != 1 {
		t.Fatalf("missing explicit API schema: %s", schema)
	}
	service, err := os.ReadFile(filepath.Join(out, "service.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(service), "NewOrderApiServiceClient") || !strings.Contains(string(service), "NewBackendServiceClient") {
		t.Fatalf("wrong backend client generation: %s", service)
	}
	mod, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "go.yorun.ai/vine "+api.DefaultGolangVineVersion) {
		t.Fatalf("unsupported Vine requirement: %s", mod)
	}
	option.VineVersion = "v0.15.6"
	if _, err := api.CompileGolang(input, option); err == nil {
		t.Fatal("accepted runtime below the minimum Vine version")
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
		writeFileForTest(t, input, "domain "+domain+"\npub data Value { id: string }\n")
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
		writeFileForTest(t, input, "domain demo.app\n"+declarations+"data Pair { first: first.user.Value second: second.user.Value }\nactor TestActor { via client {} }\napi service AppApiService { for TestActor via client method get { output Pair } }\n")
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
