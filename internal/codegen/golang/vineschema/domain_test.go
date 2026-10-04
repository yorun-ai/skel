package vineschema

import (
	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
	contractschema "go.yorun.ai/skelc/internal/schema"
	"path/filepath"
	"testing"
)

func TestVineSchemaAdaptsNormalizedSchemaProjection(t *testing.T) {
	profile := new(model.Data{
		Name: "Profile", Description: "Profile data.", Pub: true, Sensitive: true,
		Members: []*model.DataMember{new(model.DataMember{
			Name: "displayName", Description: "Display name.", Sensitive: true,
			Type: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString, Nullable: true}),
		})},
	})
	domain := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user", Description: "User domain.", Data: []*model.Data{profile},
		Services: []*model.Service{new(model.Service{
			Name: "Profiles", Pub: true, Auth: model.AuthModeAuth,
			Audiences: []*model.ActorAudience{new(model.ActorAudience{Actor: "Client", Via: string(model.ActorViaClient)})},
			Methods: []*model.Method{new(model.Method{
				Name: "get", Description: "Gets a profile.", Auth: model.AuthModeNoAuth,
				ResultType: codegentest.DataType(profile),
			})},
		})},
	})
	document, err := contractschema.Project(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	gen := newGen(Option{Domain: domain, View: mustView(t, view.ModeFull, domain), Mode: view.ModeFull})
	runtime := mustBuildDomainSchema(t, gen)

	projectedData := contractschema.Find(document, string(contractschema.DeclarationTypeData), profile.SkelName)
	if runtime.Description != document.Description || runtime.Data[0].Description != projectedData.Description {
		t.Fatalf("runtime metadata diverged from normalized schema: %#v", runtime.Data[0])
	}
	if runtime.Data[0].Sensitive != projectedData.Data.Sensitive || runtime.Data[0].Members[0].Sensitive != projectedData.Data.Members[0].Sensitive {
		t.Fatalf("runtime sensitivity diverged from normalized schema: %#v", runtime.Data[0])
	}
	projectedType := projectedData.Data.Members[0].Type
	runtimeType := runtime.Data[0].Members[0].Type
	if string(runtimeType.Kind) != string(projectedType.Kind) || string(runtimeType.Scalar) != projectedType.Name || runtimeType.Nullable != projectedType.Nullable {
		t.Fatalf("runtime type diverged from normalized schema: projected=%#v runtime=%#v", projectedType, runtimeType)
	}

	serviceValue := domain.Services()[0]
	projectedService := contractschema.Find(document, string(contractschema.DeclarationTypeService), serviceValue.SkelName)
	runtimeService := runtime.Services[0]
	if string(runtimeService.AuthMode) != string(projectedService.Service.Auth) || runtimeService.Audiences[0].SkelName != projectedService.Service.Audiences[0].Actor {
		t.Fatalf("runtime service diverged from normalized schema: projected=%#v runtime=%#v", projectedService.Service, runtimeService)
	}
	if string(runtimeService.Methods[0].AuthMode) != string(projectedService.Service.Methods[0].Auth) || runtimeService.Methods[0].Description != projectedService.Service.Methods[0].Description {
		t.Fatalf("runtime method diverged from normalized schema: projected=%#v runtime=%#v", projectedService.Service.Methods[0], runtimeService.Methods[0])
	}
}

func TestBuildDomainSchemaCopiesHashes(t *testing.T) {
	userProfile := &model.Data{
		Name: "UserProfile",
		Members: []*model.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name:        "demo.user",
		Description: "User domain",
		Data:        []*model.Data{userProfile},
		Actors: []*model.Actor{{
			Name: "ClientActor",
			Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
		}},
		Services: []*model.Service{{
			Name:      "UserService",
			Audiences: []*model.ActorAudience{{Actor: "ClientActor", Via: string(model.ActorViaClient)}},
			Methods: []*model.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(userProfile),
			}},
		}},
	})
	fillModelHashesForTest(pkg)

	gen := newGen(Option{
		CompilerVersion: "v1.2.3",
		Domain:          pkg,
		View:            mustView(t, view.ModeFull, pkg),
		Mode:            view.ModeFull,
		PackageName:     "skeled",
		Out:             filepath.Join(t.TempDir(), "skeled"),
	})
	meta := mustBuildDomainSchema(t, gen)

	if meta.Generated == nil || meta.Generated.CompilerVersion != "v1.2.3" {
		t.Fatalf("unexpected generated info: %+v", meta.Generated)
	}
	if meta.Hash != "domain-hash" {
		t.Fatalf("unexpected domain hash: %q", meta.Hash)
	}
	if !meta.Full {
		t.Fatal("expected full domain schema")
	}
	if len(meta.Data) != 1 || meta.Data[0].Hash != "data-hash" {
		t.Fatalf("unexpected data hash: %+v", meta.Data)
	}
	if len(meta.Actors) != 1 || meta.Actors[0].Hash != "actor-hash" {
		t.Fatalf("unexpected actor hash: %+v", meta.Actors)
	}
	if meta.Actors[0].AuthEnabled {
		t.Fatal("expected actor auth disabled")
	}
	if len(meta.Services) != 1 || meta.Services[0].Hash != "service-hash" {
		t.Fatalf("unexpected service hash: %+v", meta.Services)
	}
	if len(meta.Services[0].Audiences) != 1 || string(meta.Services[0].Audiences[0].Via) != string(model.ActorViaClient) {
		t.Fatalf("unexpected service for via: %+v", meta.Services[0].Audiences)
	}
	if len(meta.Services[0].Methods) != 1 || meta.Services[0].Methods[0].Hash != "method-hash" {
		t.Fatalf("unexpected method hash: %+v", meta.Services[0].Methods)
	}
	if string(meta.Services[0].AuthMode) != string(model.AuthModeRequired) {
		t.Fatalf("expected service auth required, got %s", meta.Services[0].AuthMode)
	}
	if string(meta.Services[0].Methods[0].AuthMode) != "inherit" {
		t.Fatalf("expected method auth inherit, got %s", meta.Services[0].Methods[0].AuthMode)
	}
}

func TestBuildDomainSchemaIncludesSensitiveMetadata(t *testing.T) {
	sensitiveData := &model.Data{
		Name:      "Credential",
		Sensitive: true,
		Members: []*model.DataMember{{
			Name:      "token",
			Sensitive: true,
			Type:      codegentest.StringType(),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{sensitiveData},
		Services: []*model.Service{{
			Name: "CredentialService",
			Methods: []*model.Method{{
				Name:               "exchange",
				ArgumentsSensitive: true,
				ResultSensitive:    true,
				Arguments: []*model.Argument{{
					Name:      "token",
					Sensitive: true,
					Type:      codegentest.StringType(),
				}},
				ResultType: codegentest.DataType(sensitiveData),
			}},
		}},
		Tasks: []*model.Task{{
			Name: "RotateCredentialTask",
			Triggers: []*model.TaskTrigger{{
				Name:               "manually",
				ArgumentsSensitive: true,
				Arguments: []*model.Argument{{
					Name:      "token",
					Sensitive: true,
					Type:      codegentest.StringType(),
				}},
			}},
		}},
	})

	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})
	schema := mustBuildDomainSchema(t, gen)

	if !schema.Data[0].Sensitive || !schema.Data[0].Members[0].Sensitive {
		t.Fatalf("unexpected sensitive data schema: %+v", schema.Data[0])
	}
	method := schema.Services[0].Methods[0]
	if !method.ArgumentsSensitive || !method.ResultSensitive || !method.Arguments[0].Sensitive {
		t.Fatalf("unexpected sensitive method schema: %+v", method)
	}
	trigger := schema.Tasks[0].Triggers[0]
	if !trigger.ArgumentsSensitive || !trigger.Arguments[0].Sensitive {
		t.Fatalf("unexpected sensitive trigger schema: %+v", trigger)
	}
}

func TestBuildDomainSchemaSplitFullFlagAndContent(t *testing.T) {
	pubData := &model.Data{
		Pub:  true,
		Name: "PubData",
		Members: []*model.DataMember{
			{Name: "id", Type: codegentest.StringType()},
		},
	}
	regularData := &model.Data{
		Name: "RegularData",
		Members: []*model.DataMember{
			{Name: "id", Type: codegentest.StringType()},
		},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Data: []*model.Data{pubData, regularData},
		Services: []*model.Service{
			{
				Pub:  true,
				Name: "PubService",
				Methods: []*model.Method{
					{Name: "getPub", ResultType: codegentest.DataType(pubData)},
				},
			},
			{
				Name: "RegularService",
				Methods: []*model.Method{
					{Name: "getRegular", ResultType: codegentest.DataType(regularData)},
				},
			},
		},
	})

	pubGen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModePub, pkg),
		Mode:        view.ModePub,
		PackageName: "userpub",
		Out:         filepath.Join(t.TempDir(), "pub"),
	})
	pubSchema := mustBuildDomainSchema(t, pubGen)
	if pubSchema.Full {
		t.Fatal("did not expect pub schema to be full")
	}
	if len(pubSchema.Data) != 1 || pubSchema.Data[0].Name != "PubData" {
		t.Fatalf("unexpected pub schema data: %+v", pubSchema.Data)
	}
	if len(pubSchema.Services) != 1 || pubSchema.Services[0].Name != "PubService" {
		t.Fatalf("unexpected pub schema services: %+v", pubSchema.Services)
	}

	regularGen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeRegular, pkg),
		Mode:        view.ModeRegular,
		PackageName: "user",
		Out:         filepath.Join(t.TempDir(), "regular"),
	})
	regularSchema := mustBuildDomainSchema(t, regularGen)
	if !regularSchema.Full {
		t.Fatal("expected regular schema to be full")
	}
	if len(regularSchema.Data) != 2 {
		t.Fatalf("expected regular schema to include full data, got %+v", regularSchema.Data)
	}
	if len(regularSchema.Services) != 2 {
		t.Fatalf("expected regular schema to include full services, got %+v", regularSchema.Services)
	}
}

func TestBuildDomainSchemaConfigLifecycleUsesConfValue(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Configs: []*model.Data{{
			Pub:       true,
			Name:      "UserConfig",
			Lifecycle: model.ConfigLifecycleEternal,
			Members: []*model.DataMember{
				{Name: "pageSize", Type: codegentest.IntType()},
			},
		}},
	})

	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})
	meta := mustBuildDomainSchema(t, gen)

	if len(meta.Configs) != 1 {
		t.Fatalf("expected one config, got %d", len(meta.Configs))
	}
	if meta.Configs[0].Lifecycle != "ETERNAL" {
		t.Fatalf("unexpected config lifecycle: %s", meta.Configs[0].Lifecycle)
	}
	if !meta.Configs[0].Pub {
		t.Fatal("expected config pub flag")
	}
}

func TestBuildDomainSchemaIncludesActorAuthMethod(t *testing.T) {
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{{
			Name:        "ClientActor",
			Vias:        []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
			AuthEnabled: true,
			AuthCredential: &model.Data{
				Name: "ClientActorCredential",
				Members: []*model.DataMember{
					{Name: "token", Type: codegentest.StringType()},
				},
			},
			AuthInfo: &model.Data{
				Name: "ClientActorInfo",
				Members: []*model.DataMember{
					{Name: "userId", Type: codegentest.StringType()},
				},
			},
		}},
	})
	pkg.Actors()[0].AuthService.Hash = "auth-service-hash"
	pkg.Actors()[0].AuthMethod.Hash = "auth-method-hash"

	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})
	meta := mustBuildDomainSchema(t, gen)

	if len(meta.Actors) != 1 {
		t.Fatalf("expected one actor, got %d", len(meta.Actors))
	}
	actor := meta.Actors[0]
	if !actor.AuthEnabled {
		t.Fatal("expected actor auth enabled")
	}
	if actor.AuthMethod == nil {
		t.Fatal("expected actor auth method")
	}
	if actor.AuthMethod.SkelName != "auth" || actor.AuthMethod.Hash != "auth-method-hash" {
		t.Fatalf("unexpected actor auth method: %+v", actor.AuthMethod)
	}
	if actor.AuthService == nil || len(actor.AuthService.Methods) != 1 || actor.AuthService.Methods[0].SkelName != actor.AuthMethod.SkelName {
		t.Fatalf("unexpected actor auth service: %+v", actor.AuthService)
	}
}
