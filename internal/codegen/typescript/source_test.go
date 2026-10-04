package typescript

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/sliceutil"
)

func TestNewGenDerivesPackageNameForApp(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("app"))

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.pkgName != "@yorun-ai/skeled-appapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameForAppDomain(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("sales.order"))

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.pkgName != "@yorun-ai/skeled-sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenUsesTypeScriptModule(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("sales.order"))

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		Module: "@acme/orders",
	})

	if gen.pkgName != "@acme/orders" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameFromTypeScriptModuleScope(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("sales.order"))

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.pkgName != "@acme/skeled-sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesPackageNameFromNpmScope(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("sales.order"))

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme",
	})

	if gen.pkgName != "@acme/sales-orderapi" {
		t.Fatalf("unexpected package name: %s", gen.pkgName)
	}
}

func TestNewGenDerivesExternalTypeImportsFromTypeScriptModuleScope(t *testing.T) {
	userSummary := &model.Data{Name: "UserSummary", Pub: true}
	userDomain := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{userSummary},
	})
	order := &model.Data{
		Name: "Order",
		Members: []*model.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "sales.order",
		Imports: []*model.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data:     []*model.Data{order},
		Services: []*model.Service{{Name: "OrderService", Api: true, Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{Name: "get", ResultType: codegentest.DataType(order)}}}},
	})

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.err != nil {
		t.Fatal(gen.err)
	}
	original := pkg.Data()[0].Members[0].Type
	if original.ExternalImportPath != "" {
		t.Fatal("generation mutated semantic model")
	}
	memberType := gen.bindings.Type(original)
	if memberType.ExternalImportPath != "@acme/skeled-demo-userapi" {
		t.Fatalf("unexpected import path: %s", memberType.ExternalImportPath)
	}
	if memberType.ExternalAlias != "user" {
		t.Fatalf("unexpected import alias: %s", memberType.ExternalAlias)
	}
}

func TestNewGenDerivesPublicTypeImportsFromTypeScriptModuleScope(t *testing.T) {
	userSummary := &model.Data{Name: "UserSummary", Pub: true}
	userDomain := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{userSummary},
	})
	order := &model.Data{
		Name: "Order",
		Pub:  true,
		Members: []*model.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "sales.order",
		Imports: []*model.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data: []*model.Data{order},
	})

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"), Option{
		ModuleScope: "@acme/skeled",
	})

	if gen.err != nil {
		t.Fatal(gen.err)
	}
	original := pkg.Data()[0].Members[0].Type
	if original.ExternalImportPath != "" {
		t.Fatal("generation mutated semantic model")
	}
	memberType := gen.bindings.Type(original)
	if memberType.ExternalImportPath != "@acme/skeled-demo-userapi" {
		t.Fatalf("unexpected import path: %s", memberType.ExternalImportPath)
	}
	if memberType.ExternalAlias != "user" {
		t.Fatalf("unexpected import alias: %s", memberType.ExternalAlias)
	}
}

func TestNewGenIgnoresUnusedBackendTypeImports(t *testing.T) {
	userSummary := &model.Data{Name: "UserSummary", Pub: true}
	userDomain := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{userSummary},
	})
	internalOrder := &model.Data{
		Name: "InternalOrder",
		Members: []*model.DataMember{{
			Name: "buyer",
			Type: externalDataTypeForTest(userSummary, "demo.user", "user", true),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "sales.order",
		Imports: []*model.Import{{
			Domain:        userDomain,
			Name:          "demo.user",
			Alias:         "user",
			ExplicitAlias: true,
		}},
		Data: []*model.Data{internalOrder},
	})

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"))

	if gen.err != nil {
		t.Fatal(gen.err)
	}
}

func TestRenderTsTrimsTrailingWhitespace(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user"))
	gen := newGen(pkg, outDir)

	gen.renderTs("sample.ts", "const value = 1;  \n\t\nconst next = 2;\t", nil)

	content, err := os.ReadFile(filepath.Join(outDir, "sample.ts"))
	if err != nil {
		t.Fatalf("read generated file failed: %v", err)
	}
	if got, want := string(content), "// "+common.GeneratedFileMarker+"\n\nconst value = 1;\n\nconst next = 2;\n"; got != want {
		t.Fatalf("unexpected content: got=%q want=%q", got, want)
	}
}

func TestApiViewIncludesLegacyAdmissionRulesAcrossTransports(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{
			{Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
			{Name: "AgentActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaAgent)}},
			{Name: "OpenAPIActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaOpenAPI)}},
		},
		Services: []*model.Service{
			{Name: "ClientOnlyService", Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "HybridService", Audiences: []*model.ActorAudience{{Actor: "AgentActor"}, {Actor: "ClientActor"}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "AgentOnlyService", Audiences: []*model.ActorAudience{{Actor: "AgentActor"}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "OpenAPIOnlyService", Audiences: []*model.ActorAudience{{Actor: "OpenAPIActor"}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "ClientActorOpenAPIOnlyService", Audiences: []*model.ActorAudience{{Actor: "ClientActor", Via: string(model.ActorViaOpenAPI)}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "ClientActorClientViaService", Audiences: []*model.ActorAudience{{Actor: "ClientActor", Via: string(model.ActorViaClient)}}, Methods: []*model.Method{{Name: "ping"}}},
			{Name: "InternalService", Methods: []*model.Method{{Name: "ping"}}},
		},
	})

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"))
	services := gen.apiView.Services
	got := sliceutil.Map(services, func(service *model.Service) string { return service.Name })
	if want := []string{"AgentOnlyService", "ClientActorClientViaService", "ClientActorOpenAPIOnlyService", "ClientOnlyService", "HybridService", "OpenAPIOnlyService"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected client services: got=%v want=%v", got, want)
	}
}

func TestClientServicesIncludesImportedClientActors(t *testing.T) {
	appDomain := buildModelDomainForTest(t, model.DomainSpec{
		Name:   "app",
		Actors: []*model.Actor{{Name: "UserActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}}},
	})
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Imports: []*model.Import{{
			Domain: appDomain,
			Name:   "app",
			Alias:  "app",
		}},
		Services: []*model.Service{{
			Name: "ImportedActorService", Audiences: []*model.ActorAudience{{Actor: "app.UserActor"}}, Methods: []*model.Method{{Name: "ping"}},
		}},
	})

	gen := newGen(pkg, filepath.Join(t.TempDir(), "ts"))
	services := gen.apiView.Services
	got := sliceutil.Map(services, func(service *model.Service) string { return service.Name })
	if want := []string{"ImportedActorService"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected client services: got=%v want=%v", got, want)
	}
}

func TestGenRendersTypesWithoutClientServices(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "ts")
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{{
			Name: "User", Pub: true,
			Members: []*model.DataMember{{Name: "id", Type: codegentest.IntType()}},
		}},
		Services: []*model.Service{{
			Name: "BackendService", Pub: true, Methods: []*model.Method{{Name: "ping"}},
		}},
	})

	gen := newGen(pkg, outDir)
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
