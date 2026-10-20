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
	if err := descriptor.ValidateEffectivePolicy(domain); err != nil {
		t.Fatal(err)
	}
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
	if decoded.Name != parsed.Domain.Name() || bytes.Contains(encoded, []byte(`"domain":`)) ||
		!bytes.Contains(encoded, []byte(`"name":"demo"`)) {
		t.Fatalf("descriptor lost its canonical name field: %s", encoded)
	}
	if err := descriptor.ValidateEffectivePolicy(&decoded); err != nil {
		t.Fatalf("effective policy after JSON round trip: %v", err)
	}
	method := decoded.Services[0].Methods[0]
	if method.AuthMode != descriptor.AuthModeInherit || method.EffectiveAuthMode != descriptor.AuthModeRequired ||
		method.EffectiveRequire == nil || method.EffectiveRequire.Expression.Mode != descriptor.PermissionRequireModeAll ||
		len(method.EffectiveRequire.Expression.Children) != 2 {
		t.Fatalf("descriptor lost effective policy: %+v", method)
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
	if actor.Permission == nil || actor.Permission.Service == nil || actor.Permission.Method() == nil || actor.Auth == nil || actor.Auth.Service == nil {
		t.Fatalf("descriptor lost derived actor contracts: %+v", actor)
	}
	capabilities := map[string][2]bool{
		"ClientActor": {true, true}, "GuestActor": {false, false},
		"PermissionActor": {false, true}, "TokenActor": {true, false},
	}
	if len(decoded.Actors) != len(capabilities) {
		t.Fatalf("unexpected actors: %+v", decoded.Actors)
	}
	for _, actor := range decoded.Actors {
		want, ok := capabilities[actor.Name]
		if !ok || (actor.Auth != nil) != want[0] || (actor.Permission != nil) != want[1] {
			t.Fatalf("actor %s lost declared capability presence: %+v", actor.Name, actor)
		}
		if auth := actor.Auth; auth != nil && (auth.Credential == nil || auth.Info == nil || auth.Service == nil || auth.Method() == nil) {
			t.Fatalf("actor %s lost authentication metadata: %+v", actor.Name, auth)
		}
		if permission := actor.Permission; permission != nil && (permission.Service == nil || permission.Method() == nil) {
			t.Fatalf("actor %s lost permission metadata: %+v", actor.Name, permission)
		}
		if auth := actor.Auth; auth != nil && (auth.MethodName != "auth" || auth.Method() != auth.Service.Methods[0]) {
			t.Fatalf("actor %s lost its authentication method reference: %+v", actor.Name, auth)
		}
		if permission := actor.Permission; permission != nil && (permission.MethodName != "checkCodes" || permission.Method() != permission.Service.Methods[0]) {
			t.Fatalf("actor %s lost its permission method reference: %+v", actor.Name, permission)
		}
	}
	resource := decoded.Resources[0]
	for _, current := range []*descriptor.Resource{domain.Resources[0], resource} {
		checks := []*descriptor.ResourceCheck{current.Checks[0], current.Actions[0].Checks[0], current.Actions[1].Checks[0]}
		for index, check := range checks {
			method := current.CheckMethod(check)
			if method == nil || method != current.CheckService.Methods[index] || len(method.Arguments) != 2 {
				t.Fatalf("resource check lost its canonical method: %+v", check)
			}
			if method.Arguments[0].Name != "code" || method.Arguments[1].Name != []string{"id", "ownerId", "editorId"}[index] {
				t.Fatalf("resource check resolved the wrong method: %+v", method)
			}
		}
	}
	check := decoded.Services[0].Methods[0].Require.Expression.Children[1].Check
	if check == nil || check.ServiceSkelName != resource.CheckService.SkelName ||
		check.MethodSkelName != resource.CheckMethod(resource.Checks[0]).SkelName || check.CodeArgumentName != "code" ||
		len(check.Arguments) != 1 || check.Arguments[0].Name != "id" || check.Arguments[0].JsonPath != "id" {
		t.Fatalf("descriptor lost resolved permission-check bindings: %+v", check)
	}
	element := decoded.Data[0].Members[0].Type.Element
	if element.Kind != descriptor.TypeKindData || element.SkelName != "demo.Node" {
		t.Fatalf("recursive data reference was not preserved by name: %+v", element)
	}
	for _, field := range []string{`"permission"`, `"service"`, `"methodName"`, `"expression"`} {
		if !bytes.Contains(encoded, []byte(field)) {
			t.Errorf("descriptor JSON is missing %s", field)
		}
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
			}}, Pub: true}},
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
				}},
			{
				Name: "RegularService",
				Methods: []*schema.Method{
					{Name: "getRegular", ResultType: codegentest.DataType(regularData)},
				}, Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "demo.ClientActor"}}},
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
