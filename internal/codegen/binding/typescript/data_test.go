package typescript

import (
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestCastDataRendersDeprecatedDocs(t *testing.T) {
	data := (_Types{}).castData(&schema.Data{
		Name:             "User",
		Deprecated:       true,
		DeprecatedReason: "Use Profile instead",
		Members: []*schema.DataMember{{
			Name:             "legacyId",
			Type:             codegentest.IntType(),
			Deprecated:       true,
			DeprecatedReason: "Use id instead",
		}},
	})
	output := renderTemplate(t, dataTsTemplate, &_DataTsPayload{Data: []*_Data{data}})
	if !strings.Contains(output, "* @deprecated Use Profile instead") {
		t.Fatalf("expected deprecated data docs, got:\n%s", output)
	}
	if !strings.Contains(output, "* @deprecated Use id instead") {
		t.Fatalf("expected deprecated member docs, got:\n%s", output)
	}
}

func TestBuildDataTsPayloadKeepsApiServiceDependencies(t *testing.T) {
	userStatus := &schema.Enum{Name: "UserStatus", Items: []*schema.EnumItem{{Name: "ACTIVE"}}}
	unusedStatus := &schema.Enum{Name: "UnusedStatus", Items: []*schema.EnumItem{{Name: "ACTIVE"}}}
	userProfile := &schema.Data{Name: "UserProfile"}
	user := &schema.Data{
		Name: "User",
		Members: []*schema.DataMember{
			{Name: "id", Type: codegentest.IntType()},
			{Name: "profile", Type: codegentest.DataType(userProfile)},
		},
	}
	userProfile.Members = []*schema.DataMember{{Name: "status", Type: codegentest.EnumType(userStatus)}}
	internalOnly := &schema.Data{
		Name:    "InternalOnly",
		Members: []*schema.DataMember{{Name: "status", Type: codegentest.EnumType(unusedStatus)}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name:  "demo.user",
		Enums: []*schema.Enum{userStatus, unusedStatus},
		Actors: []*schema.Actor{
			{Name: "ClientActor"},
			{Name: "AgentActor"},
		},
		Data: []*schema.Data{user, userProfile, internalOnly},
		Services: []*schema.Service{
			{Name: "ClientApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{Name: "getUser", ResultType: codegentest.DataType(user)}}},
			{Name: "AgentApiService", Api: true, AuthMode: schema.AuthModeRequired, Audiences: []*schema.ActorAudience{{Actor: "AgentActor"}}, Methods: []*schema.Method{{Name: "getInternal", ResultType: codegentest.DataType(internalOnly)}}},
		},
	})

	gen := newTestGen(pkg, ".")
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
	userStatus := &schema.Enum{Pub: true, Name: "UserStatus", Items: []*schema.EnumItem{{Name: "ACTIVE"}}}
	internalStatus := &schema.Enum{Name: "InternalStatus", Items: []*schema.EnumItem{{Name: "ACTIVE"}}}
	user := &schema.Data{
		Pub:     true,
		Name:    "User",
		Members: []*schema.DataMember{{Name: "status", Type: codegentest.EnumType(userStatus)}},
	}
	unusedPublic := &schema.Data{
		Pub:     true,
		Name:    "UnusedPublic",
		Members: []*schema.DataMember{{Name: "id", Type: codegentest.IntType()}},
	}
	internalOnly := &schema.Data{
		Name:    "InternalOnly",
		Members: []*schema.DataMember{{Name: "status", Type: codegentest.EnumType(internalStatus)}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name:  "demo.user",
		Enums: []*schema.Enum{userStatus, internalStatus},
		Data:  []*schema.Data{user, unusedPublic, internalOnly},
	})

	gen := newTestGen(pkg, ".")
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
	page := &schema.Data{
		Name:           "Page",
		TypeParameters: []*schema.TypeParameter{tItem},
		Members: []*schema.DataMember{{
			Name: "items",
			Type: codegentest.ListType(codegentest.TypeParamType(tItem)),
		}},
	}
	user := &schema.Data{
		Name:    "User",
		Members: []*schema.DataMember{{Name: "id", Type: codegentest.IntType()}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{{
			Name: "ClientActor",
		}},
		Data: []*schema.Data{page, user},
		Services: []*schema.Service{{
			Name: "ClientApiService", Api: true, AuthMode: schema.AuthModeRequired,
			Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}},
			Methods: []*schema.Method{{
				Name:       "listUsers",
				ResultType: codegentest.DataType(page, codegentest.DataType(user)),
			}},
		}},
	})

	gen := newTestGen(pkg, ".")
	payload := gen.buildDataTsPayload()

	dataNames := make([]string, 0, len(payload.Data))
	for _, data := range payload.Data {
		dataNames = append(dataNames, data.Name)
	}
	if got, want := dataNames, []string{"Page", "User"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected data: got=%v want=%v", got, want)
	}
}
