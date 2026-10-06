package descriptor

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/schema"
)

func TestBuildDomainDescriptorProducesPortableDescriptor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contract.skel")

	if err := os.WriteFile(path, []byte(portableDescriptorSource), 0o600); err != nil {
		t.Fatal(err)
	}
	parsed, err := compiler.Compile(compiler.Option{SkelIn: path})
	if err != nil {
		t.Fatal(err)
	}
	gen := newGen(Option{Domain: parsed.Domain, View: mustView(t, view.ModeFull, parsed.Domain), Mode: view.ModeFull})
	domain := gen.buildDomainDescriptor()
	encoded, err := json.Marshal(domain)
	if err != nil {
		t.Fatalf("encode descriptor of recursive schema: %v", err)
	}
	var decoded descriptor.Domain
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	roundTrip, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, roundTrip) {
		t.Fatal("descriptor metadata changed after serialization")
	}
	if decoded.Hash != parsed.Domain.Hash() || decoded.Hash == "" {
		t.Fatal("descriptor lost the compiled domain hash")
	}
	actor := decoded.Actors[0]
	if !actor.PermissionEnabled || actor.PermissionService == nil || actor.PermissionMethod == nil || actor.AuthService == nil {
		t.Fatalf("descriptor lost derived actor contracts: %+v", actor)
	}
	resource := decoded.Resources[0]
	check := decoded.Services[0].Methods[0].Require.Expression.Children[1].Check
	if check == nil || check.ServiceSkelName != resource.CheckService.SkelName ||
		check.MethodSkelName != resource.Checks[0].Method.SkelName || check.CodeArgumentName != "code" ||
		len(check.Arguments) != 1 || check.Arguments[0].Name != "id" || check.Arguments[0].JsonPath != "id" {
		t.Fatalf("descriptor lost resolved permission-check bindings: %+v", check)
	}
	element := decoded.Data[0].Members[0].Type.Element
	if element.Kind != descriptor.TypeKindData || element.SkelName != "demo.Node" {
		t.Fatalf("recursive data reference was not preserved by name: %+v", element)
	}
	for _, field := range []string{`"permissionEnabled"`, `"permissionService"`, `"permissionMethod"`, `"expression"`} {
		if !bytes.Contains(encoded, []byte(field)) {
			t.Errorf("descriptor JSON is missing %s", field)
		}
	}
}

func TestDescriptorAdaptsSemanticSchema(t *testing.T) {
	profile := new(schema.Data{
		Name: "Profile", Description: "Profile data.", Pub: true, Sensitive: true,
		Members: []*schema.DataMember{new(schema.DataMember{
			Name: "displayName", Description: "Display name.", Sensitive: true,
			Type: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString, Nullable: true}),
		})},
	})
	domain := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user", Description: "User domain.", Data: []*schema.Data{profile},
		Services: []*schema.Service{new(schema.Service{
			Name: "Profiles", Pub: true, Auth: schema.AuthModeAuth,
			Audiences: []*schema.ActorAudience{new(schema.ActorAudience{Actor: "Client", Via: string(schema.ActorViaClient)})},
			Methods: []*schema.Method{new(schema.Method{
				Name: "get", Description: "Gets a profile.", Auth: schema.AuthModeNoAuth,
				ResultType: codegentest.DataType(profile),
			})},
		})},
	})
	gen := newGen(Option{Domain: domain, View: mustView(t, view.ModeFull, domain), Mode: view.ModeFull})
	runtime := gen.buildDomainDescriptor()
	if runtime.Description != domain.Description() || runtime.Data[0].Description != profile.Description ||
		!runtime.Data[0].Sensitive || !runtime.Data[0].Members[0].Sensitive {
		t.Fatalf("descriptor lost semantic metadata: %+v", runtime.Data[0])
	}
	kind := runtime.Data[0].Members[0].Type
	if kind.Kind != descriptor.TypeKindScalar || kind.Scalar != descriptor.ScalarString || !kind.Nullable {
		t.Fatalf("invalid descriptor type: %+v", kind)
	}
	service := runtime.Services[0]
	if service.AuthMode != descriptor.AuthModeRequired || service.Audiences[0].SkelName != "demo.user.Client" ||
		service.Methods[0].AuthMode != descriptor.AuthModeOptional || service.Methods[0].Description != "Gets a profile." {
		t.Fatalf("descriptor lost normalized service policy: %+v", service)
	}
}

func TestBuildDomainDescriptorCopiesHashes(t *testing.T) {
	userProfile := &schema.Data{
		Name: "UserProfile",
		Members: []*schema.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name:        "demo.user",
		Description: "User domain",
		Data:        []*schema.Data{userProfile},
		Actors: []*schema.Actor{{
			Name: "ClientActor",
			Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
		}},
		Services: []*schema.Service{{
			Name:      "UserService",
			Audiences: []*schema.ActorAudience{{Actor: "ClientActor", Via: string(schema.ActorViaClient)}},
			Methods: []*schema.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(userProfile),
			}},
		}},
	})
	fillSchemaHashesForTest(pkg)

	gen := newGen(Option{
		CompilerVersion: "v1.2.3",
		Domain:          pkg,
		View:            mustView(t, view.ModeFull, pkg),
		Mode:            view.ModeFull,
		PackageName:     "skeled",
		Out:             filepath.Join(t.TempDir(), "skeled"),
	})
	meta := gen.buildDomainDescriptor()

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
	if len(meta.Services[0].Audiences) != 1 || string(meta.Services[0].Audiences[0].Via) != string(schema.ActorViaClient) {
		t.Fatalf("unexpected service for via: %+v", meta.Services[0].Audiences)
	}
	if len(meta.Services[0].Methods) != 1 || meta.Services[0].Methods[0].Hash != "method-hash" {
		t.Fatalf("unexpected method hash: %+v", meta.Services[0].Methods)
	}
	if string(meta.Services[0].AuthMode) != string(schema.AuthModeRequired) {
		t.Fatalf("expected service auth required, got %s", meta.Services[0].AuthMode)
	}
	if string(meta.Services[0].Methods[0].AuthMode) != "inherit" {
		t.Fatalf("expected method auth inherit, got %s", meta.Services[0].Methods[0].AuthMode)
	}
}

func TestBuildDomainDescriptorIncludesSensitiveMetadata(t *testing.T) {
	sensitiveData := &schema.Data{
		Name:      "Credential",
		Sensitive: true,
		Members: []*schema.DataMember{{
			Name:      "token",
			Sensitive: true,
			Type:      codegentest.StringType(),
		}},
	}
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{sensitiveData},
		Services: []*schema.Service{{
			Name: "CredentialService",
			Methods: []*schema.Method{{
				Name:               "exchange",
				ArgumentsSensitive: true,
				ResultSensitive:    true,
				Arguments: []*schema.Argument{{
					Name:      "token",
					Sensitive: true,
					Type:      codegentest.StringType(),
				}},
				ResultType: codegentest.DataType(sensitiveData),
			}},
		}},
		Tasks: []*schema.Task{{
			Name: "RotateCredentialTask",
			Triggers: []*schema.TaskTrigger{{
				Name:               "manually",
				ArgumentsSensitive: true,
				Arguments: []*schema.Argument{{
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
	metadata := gen.buildDomainDescriptor()

	if !metadata.Data[0].Sensitive || !metadata.Data[0].Members[0].Sensitive {
		t.Fatalf("unexpected sensitive data schema: %+v", metadata.Data[0])
	}
	method := metadata.Services[0].Methods[0]
	if !method.ArgumentsSensitive || !method.ResultSensitive || !method.Arguments[0].Sensitive {
		t.Fatalf("unexpected sensitive method schema: %+v", method)
	}
	trigger := metadata.Tasks[0].Triggers[0]
	if !trigger.ArgumentsSensitive || !trigger.Arguments[0].Sensitive {
		t.Fatalf("unexpected sensitive trigger schema: %+v", trigger)
	}
}

func TestBuildDomainDescriptorSplitFullFlagAndContent(t *testing.T) {
	pubData := &schema.Data{
		Pub:  true,
		Name: "PubData",
		Members: []*schema.DataMember{
			{Name: "id", Type: codegentest.StringType()},
		},
	}
	regularData := &schema.Data{
		Name: "RegularData",
		Members: []*schema.DataMember{
			{Name: "id", Type: codegentest.StringType()},
		},
	}
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Data: []*schema.Data{pubData, regularData},
		Services: []*schema.Service{
			{
				Pub:  true,
				Name: "PubService",
				Methods: []*schema.Method{
					{Name: "getPub", ResultType: codegentest.DataType(pubData)},
				},
			},
			{
				Name: "RegularService",
				Methods: []*schema.Method{
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
	pubSchema := pubGen.buildDomainDescriptor()
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
	regularSchema := regularGen.buildDomainDescriptor()
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

func TestBuildDomainDescriptorConfigLifecycleUsesConfValue(t *testing.T) {
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Configs: []*schema.Data{{
			Pub:       true,
			Name:      "UserConfig",
			Lifecycle: schema.ConfigLifecycleEternal,
			Members: []*schema.DataMember{
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
	meta := gen.buildDomainDescriptor()

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

func TestBuildDomainDescriptorIncludesActorAuthMethod(t *testing.T) {
	pkg := buildDescriptorDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{{
			Name:        "ClientActor",
			Vias:        []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
			AuthEnabled: true,
			AuthCredential: &schema.Data{
				Name: "ClientActorCredential",
				Members: []*schema.DataMember{
					{Name: "token", Type: codegentest.StringType()},
				},
			},
			AuthInfo: &schema.Data{
				Name: "ClientActorInfo",
				Members: []*schema.DataMember{
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
	meta := gen.buildDomainDescriptor()

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
