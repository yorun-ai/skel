package projection

import (
	"bytes"
	"reflect"
	"testing"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/schema"
)

func TestProjectMapsDeclarationKinds(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name:        "demo.contract",
		Description: "Contract domain.",
		Enums:       []*model.Enum{new(model.Enum{Name: "State", SkelName: "demo.contract.State", Items: []*model.EnumItem{}})},
		Data:        []*model.Data{new(model.Data{Name: "Record", SkelName: "demo.contract.Record", Kind: model.DataKindData, Members: []*model.DataMember{}})},
		Configs:     []*model.Data{new(model.Data{Name: "Runtime", SkelName: "demo.contract.Runtime", Kind: model.DataKindConfig, Members: []*model.DataMember{}})},
		Events:      []*model.Data{new(model.Data{Name: "Changed", SkelName: "demo.contract.Changed", Kind: model.DataKindEvent, Members: []*model.DataMember{}})},
		Actors:      []*model.Actor{new(model.Actor{Name: "Caller", SkelName: "demo.contract.Caller", Vias: []*model.ActorVia{}})},
		Resources:   []*model.Resource{new(model.Resource{Name: "Document", SkelName: "demo.contract.Document", Actions: []*model.ResourceAction{}})},
		Services:    []*model.Service{new(model.Service{Name: "Documents", SkelName: "demo.contract.Documents", Audiences: []*model.ActorAudience{}, Methods: []*model.Method{}})},
		Tasks:       []*model.Task{new(model.Task{Name: "Jobs", SkelName: "demo.contract.Jobs", Triggers: []*model.TaskTrigger{}})},
		Webs:        []*model.Web{new(model.Web{Name: "Console", SkelName: "demo.contract.Console", Audiences: []*model.ActorAudience{}})},
	})

	document, err := Project(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	if document.Domain != domain.Name() || document.Description != domain.Description() {
		t.Fatalf("document metadata = %#v", document)
	}
	want := []schema.DeclarationType{
		schema.DeclarationTypeActor, schema.DeclarationTypeConfig, schema.DeclarationTypeData, schema.DeclarationTypeEnum,
		schema.DeclarationTypeEvent, schema.DeclarationTypeResource, schema.DeclarationTypeService, schema.DeclarationTypeTask,
		schema.DeclarationTypeWeb,
	}
	got := make([]schema.DeclarationType, 0, len(document.Declarations))
	for _, declaration := range document.Declarations {
		got = append(got, declaration.Kind)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("declaration kinds = %v, want %v", got, want)
	}
}

func TestProjectMapsCompatibilityFields(t *testing.T) {
	configPos := model.Position{File: "contract.skel", Line: 2, Column: 1}
	memberPos := model.Position{File: "contract.skel", Line: 4, Column: 5}
	servicePos := model.Position{File: "contract.skel", Line: 8, Column: 1}
	methodPos := model.Position{File: "contract.skel", Line: 12, Column: 5}
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo.contract",
		Configs: []*model.Data{new(model.Data{
			Pos: configPos, Name: "Runtime", SkelName: "demo.contract.Runtime", Kind: model.DataKindConfig,
			Description: "Runtime configuration.", Deprecated: true, DeprecatedReason: "Use Settings.",
			Lifecycle: model.ConfigLifecycleInstant, Pub: true, Sensitive: true,
			TypeParameters: []*model.TypeParameter{new(model.TypeParameter{Name: "T"})},
			Members: []*model.DataMember{new(model.DataMember{
				Pos: memberPos, Name: "endpoint", Description: "Service endpoint.", Example: `"https://example.com"`,
				Sensitive: true, Type: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString, Nullable: true}),
			})},
		})},
		Services: []*model.Service{new(model.Service{
			Pos: servicePos, Name: "Documents", SkelName: "demo.contract.Documents", Pub: true,
			Description: "Document operations.", Auth: model.AuthModeNoAuth,
			Audiences: []*model.ActorAudience{new(model.ActorAudience{Actor: "Caller", Via: "client", Pos: servicePos})},
			Require: new(model.PermissionRequire{Expr: new(model.PermissionExpr{
				Mode: model.PermissionRequireModeCode, Code: "demo.contract.Document:read",
			})}),
			Methods: []*model.Method{new(model.Method{
				Pos: methodPos, Name: "get", SkelName: "get", Description: "Gets a document.", Example: "get example",
				Auth: model.AuthModeAuth, InputDescription: "Input.", ArgumentsSensitive: true,
				OutputDescription: "Output.", OutputExample: "output example", ResultSensitive: true,
				Arguments: []*model.Argument{new(model.Argument{
					Pos: methodPos, Name: "id", Description: "Document ID.", Example: `"doc-1"`, Sensitive: true,
					Type: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}),
				})},
				ResultType: new(model.Type{Kind: model.TypeKindUnresolvedReference, SkelName: "Record"}),
			})},
		})},
	})

	document, err := Project(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	config := schema.Find(document, string(schema.DeclarationTypeConfig), "demo.contract.Runtime")
	if config == nil || config.Pos != configPos || config.Data.Lifecycle != schema.ConfigLifecycleInstant || !config.Data.Sensitive {
		t.Fatalf("config projection = %#v", config)
	}
	if config.Description != "Runtime configuration." || !config.Deprecated || config.DeprecatedReason != "Use Settings." || !config.Pub {
		t.Fatalf("config metadata projection = %#v", config)
	}
	member := config.Data.Members[0]
	if member.Pos != memberPos || member.Name != "endpoint" || member.Description != "Service endpoint." || member.Type.Name != "string" || !member.Type.Nullable {
		t.Fatalf("member projection = %#v", member)
	}

	service := schema.Find(document, string(schema.DeclarationTypeService), "demo.contract.Documents")
	if service == nil || service.Pos != servicePos || service.Service.Auth != schema.AuthModeOptional {
		t.Fatalf("service projection = %#v", service)
	}
	if !reflect.DeepEqual(service.Service.Audiences, []*schema.Audience{new(schema.Audience{Actor: "demo.contract.Caller", Via: "client", Pos: servicePos})}) {
		t.Fatalf("audience projection = %#v", service.Service.Audiences)
	}
	if service.Service.Require.Mode != schema.RequirementModeCode || service.Service.Require.Code != "demo.contract.Document:read" {
		t.Fatalf("requirement projection = %#v", service.Service.Require)
	}
	method := service.Service.Methods[0]
	if method.Pos != methodPos || method.Auth != schema.AuthModeRequired || method.Result.Kind != schema.TypeKindImportedReference || method.Result.Name != "demo.contract.Record" {
		t.Fatalf("method projection = %#v", method)
	}
	if len(method.Arguments) != 1 || method.Arguments[0].Type.Name != "string" || !method.Arguments[0].Sensitive {
		t.Fatalf("argument projection = %#v", method.Arguments)
	}
}

func TestProjectMapsModelEnums(t *testing.T) {
	t.Run("config lifecycle", func(t *testing.T) {
		tests := []struct {
			semantic model.ConfigLifecycle
			wire     schema.ConfigLifecycle
		}{
			{semantic: model.ConfigLifecycleEternal, wire: schema.ConfigLifecycleEternal},
			{semantic: model.ConfigLifecycleInstant, wire: schema.ConfigLifecycleInstant},
		}
		for _, test := range tests {
			got := projectDataSchema(new(model.Data{Lifecycle: test.semantic, Members: []*model.DataMember{}})).Lifecycle
			if got != test.wire {
				t.Fatalf("lifecycle %q projects to %q, want %q", test.semantic, got, test.wire)
			}
		}
	})

	t.Run("authentication", func(t *testing.T) {
		tests := []struct {
			semantic model.AuthMode
			wire     schema.AuthMode
		}{
			{semantic: model.AuthModeUnset, wire: schema.AuthModeInherit},
			{semantic: model.AuthModeAuth, wire: schema.AuthModeRequired},
			{semantic: model.AuthModeNoAuth, wire: schema.AuthModeOptional},
		}
		for _, test := range tests {
			if got := normalizedAuth(test.semantic); got != test.wire {
				t.Fatalf("auth %q projects to %q, want %q", test.semantic, got, test.wire)
			}
		}
	})
}

func TestProjectMapsTypeKinds(t *testing.T) {
	tests := []struct {
		name     string
		semantic *model.Type
		wire     *schema.Type
	}{
		{name: "imported reference", semantic: new(model.Type{Kind: model.TypeKindUnresolvedReference, SkelName: "User", ExternalAlias: "shared"}), wire: new(schema.Type{Kind: schema.TypeKindImportedReference, Name: "shared.User"})},
		{name: "scalar", semantic: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarBoolean}), wire: new(schema.Type{Kind: schema.TypeKindScalar, Name: "bool"})},
		{name: "list", semantic: new(model.Type{Kind: model.TypeKindList, List: new(model.ListType{Value: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString})})}), wire: new(schema.Type{Kind: schema.TypeKindList, Element: new(schema.Type{Kind: schema.TypeKindScalar, Name: "string"})})},
		{name: "map", semantic: new(model.Type{Kind: model.TypeKindMap, Map: new(model.MapType{Key: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString}), Value: new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarInt})})}), wire: new(schema.Type{Kind: schema.TypeKindMap, Key: new(schema.Type{Kind: schema.TypeKindScalar, Name: "string"}), Value: new(schema.Type{Kind: schema.TypeKindScalar, Name: "int"})})},
		{name: "enum", semantic: new(model.Type{Kind: model.TypeKindEnum, SkelName: "demo.contract.State"}), wire: new(schema.Type{Kind: schema.TypeKindEnum, Name: "demo.contract.State"})},
		{name: "data", semantic: new(model.Type{Kind: model.TypeKindData, SkelName: "demo.contract.Page", Nullable: true, TypeArguments: []*model.Type{new(model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarString})}}), wire: new(schema.Type{Kind: schema.TypeKindData, Nullable: true, Name: "demo.contract.Page", Arguments: []*schema.Type{new(schema.Type{Kind: schema.TypeKindScalar, Name: "string"})}})},
		{name: "config", semantic: new(model.Type{Kind: model.TypeKindData, Data: new(model.Data{Kind: model.DataKindConfig}), SkelName: "demo.contract.RuntimeConfig"}), wire: new(schema.Type{Kind: schema.TypeKindConfig, Name: "demo.contract.RuntimeConfig"})},
		{name: "event", semantic: new(model.Type{Kind: model.TypeKindData, Data: new(model.Data{Kind: model.DataKindEvent}), SkelName: "demo.contract.ChangedEvent"}), wire: new(schema.Type{Kind: schema.TypeKindEvent, Name: "demo.contract.ChangedEvent"})},
		{name: "type parameter", semantic: new(model.Type{Kind: model.TypeKindTypeParameter, TypeParameter: new(model.TypeParameter{Name: "T"})}), wire: new(schema.Type{Kind: schema.TypeKindTypeParameter, Name: "T"})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := projectType(test.semantic); !reflect.DeepEqual(got, test.wire) {
				t.Fatalf("type projection mismatch:\nwant: %#v\ngot:  %#v", test.wire, got)
			}
		})
	}
}

func TestProjectMapsRequirementModes(t *testing.T) {
	check := new(model.PermissionCheckInvocation{ResourceSkelName: "demo.contract.Document", CheckName: "owns"})
	tests := []struct {
		name     string
		semantic *model.PermissionExpr
		wire     schema.RequirementMode
	}{
		{name: "code", semantic: new(model.PermissionExpr{Mode: model.PermissionRequireModeCode, Code: "demo.contract.Document:read"}), wire: schema.RequirementModeCode},
		{name: "check", semantic: new(model.PermissionExpr{Mode: model.PermissionRequireModeCheck, Check: check}), wire: schema.RequirementModeCheck},
		{name: "all", semantic: new(model.PermissionExpr{Mode: model.PermissionRequireModeAll, Children: []*model.PermissionExpr{}}), wire: schema.RequirementModeAll},
		{name: "any", semantic: new(model.PermissionExpr{Mode: model.PermissionRequireModeAny, Children: []*model.PermissionExpr{}}), wire: schema.RequirementModeAny},
		{name: "reference", semantic: new(model.PermissionExpr{Check: check}), wire: schema.RequirementModeReference},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := projectRequirementExpr(test.semantic); got.Mode != test.wire {
				t.Fatalf("requirement mode = %q, want %q", got.Mode, test.wire)
			}
		})
	}
}

func TestActorIdentifierProjectionAndDiff(t *testing.T) {
	actor := &model.Actor{Name: "UserActor", SkelName: "demo.UserActor", AuthEnabled: true, Vias: []*model.ActorVia{{Name: "client"}},
		AuthCredential: &model.Data{Members: []*model.DataMember{}},
		AuthInfo:       &model.Data{Members: []*model.DataMember{{Name: "id", Type: &model.Type{Kind: model.TypeKindScalar, Scalar: model.ScalarInt}}}},
	}
	domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Actors: []*model.Actor{actor}})
	baseline, err := Project(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	actor.IdentifierField = "id"
	candidate, err := Project(domain, nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := schema.Encode(&buf, candidate); err != nil {
		t.Fatal(err)
	}
	decoded, err := schema.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Declarations[0].Actor.IdentifierField != "id" {
		t.Fatal("lost identifier field")
	}
	for _, pair := range [][2]*schema.Document{{baseline, decoded}, {decoded, baseline}} {
		report, err := schema.Diff(pair[0], pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Changes) != 1 || report.Changes[0].Code != "actor.identifier.changed" || report.Changes[0].Impact != schema.ImpactDangerous {
			t.Fatalf("unexpected diff: %+v", report.Changes)
		}
	}
	for _, name := range []string{"missing", "id"} {
		decoded.Declarations[0].Actor.IdentifierField = name
		decoded.Declarations[0].Actor.AuthInfo.Members[0].Type.Nullable = true
		if err := schema.Validate(decoded); err == nil {
			t.Fatal("accepted invalid identifier")
		}
	}
}

func TestWebMountProjectionRoundTrip(t *testing.T) {
	for _, path := range []string{"", "/", "/portal/v1-assets/"} {
		domain := model.NewDomainFromSpec(model.DomainSpec{Name: "demo", Webs: []*model.Web{
			{Name: "PortalWeb", SkelName: "demo.PortalWeb", MountPath: path, Audiences: []*model.ActorAudience{}},
		}})
		document, err := Project(domain, nil)
		if err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if err := schema.Encode(&output, document); err != nil {
			t.Fatal(err)
		}
		decoded, err := schema.Decode(&output)
		if err != nil {
			t.Fatal(err)
		}
		if decoded.Declarations[0].Web.MountPath != path {
			t.Fatalf("mount path lost: %+v", decoded)
		}
	}
}
