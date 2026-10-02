package source

import (
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/codegen/codegentest"
	"go.yorun.ai/skelc/internal/model"
)

func TestCastData(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name:        "Page",
		Description: "Paginated result",
		TypeParameters: []*model.TypeParameter{
			{Name: "TItem"},
		},
		Members: []*model.DataMember{
			{
				Name:        "generatedAt",
				Description: "Generated at",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarTimestamp,
				},
			},
			{
				Name:        "avatarUrl",
				Description: "Avatar URL",
				Example:     `"https://xxx.com/a.png"`,
				Type: &model.Type{
					Kind:     model.TypeKindScalar,
					Scalar:   model.ScalarString,
					Nullable: true,
				},
			},
		},
	})

	if data.FullName != "Page<TItem>" {
		t.Fatalf("unexpected full name: %s", data.FullName)
	}
	if len(data.CommentLines) == 0 || data.CommentLines[0] != "Paginated result." {
		t.Fatalf("unexpected data comment lines: %+v", data.CommentLines)
	}
	if len(data.Members) != 2 {
		t.Fatalf("unexpected member count: %d", len(data.Members))
	}
	if data.Members[0].Type.Plain != "string" {
		t.Fatalf("unexpected first member type: %s", data.Members[0].Type.Plain)
	}
	if len(data.Members[1].CommentLines) == 0 || data.Members[1].CommentLines[0] != `Avatar URL (e.g. "https://xxx.com/a.png").` {
		t.Fatalf("unexpected second member comment lines: %+v", data.Members[1].CommentLines)
	}
}

func TestCastDataRendersDeprecatedDocs(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name:             "User",
		Deprecated:       true,
		DeprecatedReason: "Use Profile instead",
		Members: []*model.DataMember{{
			Name:             "legacyId",
			Type:             codegentest.IntType(),
			Deprecated:       true,
			DeprecatedReason: "Use id instead",
		}},
	})
	output := renderTemplate(t, dataTsTemplate, &DataTsPayload{Data: []*Data{data}})
	if !strings.Contains(output, "* @deprecated Use Profile instead") {
		t.Fatalf("expected deprecated data docs, got:\n%s", output)
	}
	if !strings.Contains(output, "* @deprecated Use id instead") {
		t.Fatalf("expected deprecated member docs, got:\n%s", output)
	}
}

func TestCastDataMapsDurationToString(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name: "TimeoutConfig",
		Members: []*model.DataMember{
			{
				Name: "timeout",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarDuration,
				},
			},
		},
	})
	if data.Members[0].Type.Plain != "string" {
		t.Fatalf("unexpected duration member type: %s", data.Members[0].Type.Plain)
	}
}

func TestCastDataMapsLocalDateToString(t *testing.T) {
	data := (_Types{}).castData(&model.Data{
		Name: "Profile",
		Members: []*model.DataMember{
			{
				Name: "birthday",
				Type: &model.Type{
					Kind:   model.TypeKindScalar,
					Scalar: model.ScalarLocalDate,
				},
			},
		},
	})
	if data.Members[0].Type.Plain != "string" {
		t.Fatalf("unexpected date member type: %s", data.Members[0].Type.Plain)
	}
}

func TestTypesTemplateKeepsModuleSemanticsWhenEmpty(t *testing.T) {
	output := renderTemplate(t, dataTsTemplate, &DataTsPayload{})
	if !strings.Contains(output, "export {};") {
		t.Fatalf("expected rendered types to keep module semantics, got:\n%s", output)
	}
}

func TestBuildDataTsPayloadKeepsLegacyServiceDependencies(t *testing.T) {
	userStatus := &model.Enum{Name: "UserStatus", Items: []*model.EnumItem{{Name: "ACTIVE"}}}
	unusedStatus := &model.Enum{Name: "UnusedStatus", Items: []*model.EnumItem{{Name: "ACTIVE"}}}
	userProfile := &model.Data{Name: "UserProfile"}
	user := &model.Data{
		Name: "User",
		Members: []*model.DataMember{
			{Name: "id", Type: codegentest.IntType()},
			{Name: "profile", Type: codegentest.DataType(userProfile)},
		},
	}
	userProfile.Members = []*model.DataMember{{Name: "status", Type: codegentest.EnumType(userStatus)}}
	internalOnly := &model.Data{
		Name:    "InternalOnly",
		Members: []*model.DataMember{{Name: "status", Type: codegentest.EnumType(unusedStatus)}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name:  "demo.user",
		Enums: []*model.Enum{userStatus, unusedStatus},
		Actors: []*model.Actor{
			{Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
			{Name: "AgentActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaAgent)}},
		},
		Data: []*model.Data{user, userProfile, internalOnly},
		Services: []*model.Service{
			{Name: "ClientService", Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{Name: "getUser", ResultType: codegentest.DataType(user)}}},
			{Name: "AgentService", Audiences: []*model.ActorAudience{{Actor: "AgentActor"}}, Methods: []*model.Method{{Name: "getInternal", ResultType: codegentest.DataType(internalOnly)}}},
		},
	})

	gen := newGen(pkg, ".")
	payload := gen.buildDataTsPayload()

	enumNames := make([]string, 0, len(payload.Enums))
	for _, enum := range payload.Enums {
		enumNames = append(enumNames, enum.Name)
	}
	if got, want := enumNames, []string{"UnusedStatus", "UserStatus"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected enums: got=%v want=%v", got, want)
	}
	dataNames := make([]string, 0, len(payload.Data))
	for _, data := range payload.Data {
		dataNames = append(dataNames, data.Name)
	}
	if got, want := dataNames, []string{"InternalOnly", "User", "UserProfile"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected data: got=%v want=%v", got, want)
	}
}

func TestBuildDataTsPayloadKeepsExplicitPubTypes(t *testing.T) {
	userStatus := &model.Enum{Pub: true, Name: "UserStatus", Items: []*model.EnumItem{{Name: "ACTIVE"}}}
	internalStatus := &model.Enum{Name: "InternalStatus", Items: []*model.EnumItem{{Name: "ACTIVE"}}}
	user := &model.Data{
		Pub:     true,
		Name:    "User",
		Members: []*model.DataMember{{Name: "status", Type: codegentest.EnumType(userStatus)}},
	}
	unusedPublic := &model.Data{
		Pub:     true,
		Name:    "UnusedPublic",
		Members: []*model.DataMember{{Name: "id", Type: codegentest.IntType()}},
	}
	internalOnly := &model.Data{
		Name:    "InternalOnly",
		Members: []*model.DataMember{{Name: "status", Type: codegentest.EnumType(internalStatus)}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name:  "demo.user",
		Enums: []*model.Enum{userStatus, internalStatus},
		Data:  []*model.Data{user, unusedPublic, internalOnly},
	})

	gen := newGen(pkg, ".")
	payload := gen.buildDataTsPayload()

	enumNames := make([]string, 0, len(payload.Enums))
	for _, enum := range payload.Enums {
		enumNames = append(enumNames, enum.Name)
	}
	if got, want := enumNames, []string{"UserStatus"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected enums: got=%v want=%v", got, want)
	}
	dataNames := make([]string, 0, len(payload.Data))
	for _, data := range payload.Data {
		dataNames = append(dataNames, data.Name)
	}
	if got, want := dataNames, []string{"UnusedPublic", "User"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected data: got=%v want=%v", got, want)
	}
}

func TestBuildDataTsPayloadKeepsGenericTypeArguments(t *testing.T) {
	tItem := codegentest.TypeParam("TItem")
	page := &model.Data{
		Name:           "Page",
		TypeParameters: []*model.TypeParameter{tItem},
		Members: []*model.DataMember{{
			Name: "items",
			Type: codegentest.ListType(codegentest.TypeParamType(tItem)),
		}},
	}
	user := &model.Data{
		Name:    "User",
		Members: []*model.DataMember{{Name: "id", Type: codegentest.IntType()}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{{
			Name: "ClientActor",
			Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
		}},
		Data: []*model.Data{page, user},
		Services: []*model.Service{{
			Name:      "ClientService",
			Audiences: []*model.ActorAudience{{Actor: "ClientActor"}},
			Methods: []*model.Method{{
				Name:       "listUsers",
				ResultType: codegentest.DataType(page, codegentest.DataType(user)),
			}},
		}},
	})

	gen := newGen(pkg, ".")
	payload := gen.buildDataTsPayload()

	dataNames := make([]string, 0, len(payload.Data))
	for _, data := range payload.Data {
		dataNames = append(dataNames, data.Name)
	}
	if got, want := dataNames, []string{"Page", "User"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected data: got=%v want=%v", got, want)
	}
}
