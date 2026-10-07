package typescript

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/output"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

func TestNewGenDerivesPackageNameForApp(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("app"))

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.pkgName != "@yorun-ai/skeled-appapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameForAppDomain(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("sales.order"))

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.pkgName != "@yorun-ai/skeled-sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenUsesTypeScriptModule(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("sales.order"))

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		Module: "@acme/orders",
	})

	if gen.pkgName != "@acme/orders" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameFromTypeScriptModuleScope(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("sales.order"))

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.pkgName != "@acme/skeled-sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameFromNpmScope(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("sales.order"))

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme",
	})

	if gen.pkgName != "@acme/sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesExternalTypeImportsFromTypeScriptModuleScope(t *testing.T) {
	userSummary := &schema.Data{Name: "UserSummary", Pub: true}
	userDomain := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{userSummary},
	})
	order := &schema.Data{
		Name: "Order",
		Members: []*schema.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "sales.order",
		Imports: []*schema.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data:     []*schema.Data{order},
		Services: []*schema.Service{{Name: "OrderService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{Name: "get", ResultType: codegentest.DataType(order)}}}},
	})

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.err != nil {
		t.Fatal(gen.err)
	}
	original := pkg.Data()[0].Members[0].Type
	if original.ExternalAlias != "user" {
		t.Fatal("generation mutated semantic schema")
	}
	memberType := gen.bindings[original]
	if memberType.Path != "@acme/skeled-demo-userapi" {
		t.Fatalf("unexpected import path: %s", memberType.Path)
	}
	if memberType.Alias != "user" {
		t.Fatalf("unexpected import alias: %s", memberType.Alias)
	}
}

func TestNewGenDerivesPublicTypeImportsFromTypeScriptModuleScope(t *testing.T) {
	userSummary := &schema.Data{Name: "UserSummary", Pub: true}
	userDomain := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{userSummary},
	})
	order := &schema.Data{
		Name: "Order",
		Pub:  true,
		Members: []*schema.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "sales.order",
		Imports: []*schema.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data: []*schema.Data{order},
	})

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.err != nil {
		t.Fatal(gen.err)
	}
	original := pkg.Data()[0].Members[0].Type
	if original.ExternalAlias != "user" {
		t.Fatal("generation mutated semantic schema")
	}
	memberType := gen.bindings[original]
	if memberType.Path != "@acme/skeled-demo-userapi" {
		t.Fatalf("unexpected import path: %s", memberType.Path)
	}
	if memberType.Alias != "user" {
		t.Fatalf("unexpected import alias: %s", memberType.Alias)
	}
}

func TestNewGenIgnoresUnusedBackendTypeImports(t *testing.T) {
	userSummary := &schema.Data{Name: "UserSummary", Pub: true}
	userDomain := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{userSummary},
	})
	internalOrder := &schema.Data{
		Name: "InternalOrder",
		Members: []*schema.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "sales.order",
		Imports: []*schema.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data: []*schema.Data{internalOrder},
	})

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.err != nil {
		t.Fatal(gen.err)
	}
}

func TestRenderTsTrimsTrailingWhitespace(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	pkg := buildSchemaDomainForTest(t, codegentest.DomainSchema("demo.user"))
	gen := newTestGen(pkg, outDir)

	gen.renderTs("sample.ts", "const value = 1;  \n\t\nconst next = 2;\t", nil)

	content, err := os.ReadFile(filepath.Join(outDir, "sample.ts"))
	if err != nil {
		t.Fatalf("read generated file failed: %v", err)
	}
	if got, want := string(content), "// "+output.GeneratedFileMarker+"\n\nconst value = 1;\n\nconst next = 2;\n"; got != want {
		t.Fatalf("unexpected content: got=%q want=%q", got, want)
	}
}

func TestApiViewIncludesExplicitServicesAcrossTransports(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{
			{Name: "ClientActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}},
			{Name: "AgentActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaAgent)}},
			{Name: "OpenAPIActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaOpenAPI)}},
		},
		Services: []*schema.Service{
			{Name: "ClientOnlyApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "HybridApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "AgentActor"}, {Actor: "ClientActor"}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "AgentOnlyApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "AgentActor"}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "OpenAPIOnlyApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "OpenAPIActor"}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "ClientActorOpenAPIOnlyApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "ClientActor", Via: string(schema.ActorViaOpenAPI)}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "ClientActorClientViaApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "ClientActor", Via: string(schema.ActorViaClient)}}, Methods: []*schema.Method{{Name: "ping"}}},
			{Name: "InternalService", Pub: true, Methods: []*schema.Method{{Name: "ping"}}},
		},
	})

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"))
	services := gen.apiView.Services
	got := sliceutil.Map(services, func(service *schema.Service) string { return service.Name })
	if want := []string{"AgentOnlyApiService", "ClientActorClientViaApiService", "ClientActorOpenAPIOnlyApiService", "ClientOnlyApiService", "HybridApiService", "OpenAPIOnlyApiService"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected client services: got=%v want=%v", got, want)
	}
}

func TestClientServicesIncludesImportedClientActors(t *testing.T) {
	appDomain := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name:   "app",
		Actors: []*schema.Actor{{Name: "UserActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}}},
	})
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Imports: []*schema.Import{{
			Domain: appDomain,
			Name:   "app",
			Alias:  "app",
		}},
		Services: []*schema.Service{{
			Name: "ImportedActorApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "app.UserActor"}}, Methods: []*schema.Method{{Name: "ping"}},
		}},
	})

	gen := newTestGen(pkg, filepath.Join(t.TempDir(), "ts"))
	services := gen.apiView.Services
	got := sliceutil.Map(services, func(service *schema.Service) string { return service.Name })
	if want := []string{"ImportedActorApiService"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected client services: got=%v want=%v", got, want)
	}
}

func TestGenRendersTypesWithoutClientServices(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{{
			Name: "User", Pub: true,
			Members: []*schema.DataMember{{Name: "id", Type: codegentest.IntType()}},
		}},
		Services: []*schema.Service{{
			Name: "BackendService", Pub: true, Methods: []*schema.Method{{Name: "ping"}},
		}},
	})

	gen := newTestGen(pkg, outDir)
	if gen.err != nil {
		t.Fatal(gen.err)
	}
	if len(gen.apiView.Services) != 0 {
		t.Fatalf("fixture must not expose client services: %+v", gen.apiView.Services)
	}
	gen.generate()
	if err := gen.renderer.Err(); err != nil {
		t.Fatal(err)
	}

	for _, filename := range []string{indexFilename, dataTsFilename, serviceTsFilename, specTsFilename} {
		content, err := os.ReadFile(filepath.Join(outDir, filename))
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		if strings.Contains(string(content), "BackendService") {
			t.Fatalf("backend service leaked into %s: %s", filename, content)
		}
		if filename == dataTsFilename && !strings.Contains(string(content), "export type User =") {
			t.Fatalf("public type missing from data.ts: %s", content)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "package.json")); !os.IsNotExist(err) {
		t.Fatalf("expected package.json to be missing, err=%v", err)
	}
}
