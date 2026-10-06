package typescript

import (
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestBuildServiceNames(t *testing.T) {
	names := buildServiceNames("userService")

	if names.Name != "UserService" {
		t.Fatalf("unexpected service name: %s", names.Name)
	}
	if names.FactoryName != "createUserService" {
		t.Fatalf("unexpected factory name: %s", names.FactoryName)
	}
	if names.SpecName != "UserServiceSpec" {
		t.Fatalf("unexpected spec name: %s", names.SpecName)
	}
}

func TestCastService(t *testing.T) {
	service := (_Types{}).castService(&schema.Service{
		Name:        "UserService",
		SkelName:    "demo.user.UserService",
		Description: "User service",
		Methods: []*schema.Method{
			{
				Name:        "getUser",
				Description: "Get a user by ID",
				Arguments: []*schema.Argument{
					{
						Name:        "userId",
						Description: "User ID",
						Example:     `"10001"`,
						Type: &schema.Type{
							Kind:   schema.TypeKindScalar,
							Scalar: schema.ScalarInt,
						},
					},
				},
				ResultType: &schema.Type{
					Kind: schema.TypeKindData,
					Data: &schema.Data{
						Name: "User",
					},
					Nullable: true,
				},
				OutputDescription: "User information",
				OutputExample:     `{ id:10001, name:"zhangsan" }`,
				ArgumentsData: &schema.Data{
					Name: "UserServiceGetUserArguments",
					Members: []*schema.DataMember{
						{
							Name: "userId",
							Type: &schema.Type{
								Kind:   schema.TypeKindScalar,
								Scalar: schema.ScalarInt,
							},
						},
					},
				},
			},
		},
	})

	if len(service.CommentLines) == 0 || service.CommentLines[0] != "User service" {
		t.Fatalf("unexpected service comment lines: %+v", service.CommentLines)
	}
	if len(service.Methods) != 1 {
		t.Fatalf("unexpected method count: %d", len(service.Methods))
	}
	if service.SpecName != "UserServiceSpec" {
		t.Fatalf("unexpected spec name: %s", service.SpecName)
	}
	if got := service.Methods[0].SummaryLines; !reflect.DeepEqual(got, []string{"Get a user by ID."}) {
		t.Fatalf("unexpected method summary lines: %+v", got)
	}
	if got := service.Methods[0].ParamDocs; !reflect.DeepEqual(got, []*_MethodParamDoc{
		{Name: "params", Description: "Request parameters"},
		{Name: "options", Description: "Call options, optional"},
	}) {
		t.Fatalf("unexpected method param docs: %+v", got)
	}
	if service.Methods[0].ReturnDoc == nil || service.Methods[0].ReturnDoc.Description != `User information (e.g. { id:10001, name:"zhangsan" })` {
		t.Fatalf("unexpected method return doc: %+v", service.Methods[0].ReturnDoc)
	}
	if !service.Methods[0].HasParams {
		t.Fatal("expected method params")
	}
	if service.Methods[0].Arguments[0].Name != "userId" {
		t.Fatalf("unexpected argument name: %s", service.Methods[0].Arguments[0].Name)
	}
}

func TestServiceTemplateRendersDeprecatedDocs(t *testing.T) {
	service := (_Types{}).castService(&schema.Service{
		Name:             "UserService",
		Deprecated:       true,
		DeprecatedReason: "Use ProfileService instead\nComplete migration first",
		Methods: []*schema.Method{{
			Name:             "getUser",
			Deprecated:       true,
			DeprecatedReason: "Use getProfile instead",
			Arguments: []*schema.Argument{{
				Name:             "legacyId",
				Type:             codegentest.IntType(),
				Deprecated:       true,
				DeprecatedReason: "Use id instead\nLegacy IDs will be removed",
			}},
		}},
	})
	output := renderTemplate(t, serviceTsTemplate, &_ServiceTsPayload{Services: []*_Service{service}})
	for _, expected := range []string{
		"* @deprecated Use ProfileService instead",
		"* Complete migration first",
		"* @deprecated Use getProfile instead",
		"* @deprecated Use id instead",
		"* Legacy IDs will be removed",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in generated service:\n%s", expected, output)
		}
	}
}

func TestBuildServiceTypeImports(t *testing.T) {
	imports := (_Types{}).buildServiceTypeImports([]*schema.Service{{
		Methods: []*schema.Method{{
			Arguments: []*schema.Argument{{
				Type: &schema.Type{
					Kind: schema.TypeKindData,
					Data: &schema.Data{
						Name: "Page",
					},
					TypeArguments: []*schema.Type{{
						Kind: schema.TypeKindEnum,
						Enum: &schema.Enum{Name: "UserStatus"},
					}},
				},
			}},
			ResultType: &schema.Type{
				Kind: schema.TypeKindList,
				List: &schema.ListType{Value: &schema.Type{
					Kind: schema.TypeKindData,
					Data: &schema.Data{
						Name: "User",
					},
				}},
			},
		}},
	}})

	if got, want := imports, []string{"User", "Page", "UserStatus"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected imports: got=%v want=%v", got, want)
	}
}

func TestBuildServiceImportsSkipsExternalTypes(t *testing.T) {
	services := []*schema.Service{{
		Methods: []*schema.Method{{
			Arguments: []*schema.Argument{{
				Type: &schema.Type{
					Kind:          schema.TypeKindData,
					Data:          &schema.Data{Name: "UserSummary"},
					ExternalAlias: "userpub",
				},
			}},
		}},
	}}

	types := _Types{bindings: binding.TypeBindings{services[0].Methods[0].Arguments[0].Type: {Alias: "userpub", Path: "@acme/skeled-userpub"}}}
	imports := types.buildServiceTypeImports(services)
	if len(imports) != 0 {
		t.Fatalf("unexpected local imports: %+v", imports)
	}
	externalImports := types.buildServiceExternalTypeImports(services)
	if got, want := externalImports, []*_TypeImport{{Alias: "userpub", Path: "@acme/skeled-userpub"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected external imports: got=%+v want=%+v", got, want)
	}
}

func TestTypesTemplateRendersExternalImports(t *testing.T) {
	payload := &_DataTsPayload{
		TypeImports: []*_TypeImport{{Alias: "userpub", Path: "@acme/skeled-userpub"}},
		Data: []*_Data{{
			Name:     "Loan",
			FullName: "Loan",
			Members: []*_DataMember{{
				Name: "borrower",
				Type: &_Type{Plain: "userpub.UserSummary"},
			}},
		}},
	}

	output := renderTemplate(t, dataTsTemplate, payload)
	if !strings.Contains(output, "import type * as userpub from '@acme/skeled-userpub';") {
		t.Fatalf("expected external type import, got:\n%s", output)
	}
	if !strings.Contains(output, "borrower: userpub.UserSummary;") {
		t.Fatalf("expected external type reference, got:\n%s", output)
	}
}

func TestBuildServiceTsPayloadIncludesLegacyAdmissionRules(t *testing.T) {
	user := &schema.Data{
		Name: "User",
		Members: []*schema.DataMember{{
			Name: "id",
			Type: codegentest.IntType(),
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{
			{Name: "ClientActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}},
			{Name: "AgentActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaAgent)}},
		},
		Data: []*schema.Data{user},
		Services: []*schema.Service{
			{Name: "ClientService", Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
			{Name: "AgentService", Audiences: []*schema.ActorAudience{{Actor: "AgentActor"}}, Methods: []*schema.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
		},
	})

	gen := newTestGen(pkg, ".")
	payload := gen.buildServiceTsPayload()
	if len(payload.Services) != 2 {
		t.Fatalf("unexpected service count: %d", len(payload.Services))
	}
	if payload.Services[0].Name != "AgentService" || payload.Services[1].Name != "ClientService" {
		t.Fatalf("unexpected service: %s", payload.Services[0].Name)
	}
	if got, want := payload.TypeImports, []string{"User"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected type imports: got=%v want=%v", got, want)
	}
}

func TestBuildServiceTsPayloadExcludesBackendServices(t *testing.T) {
	user := &schema.Data{
		Pub:  true,
		Name: "User",
		Members: []*schema.DataMember{{
			Name: "id",
			Type: codegentest.IntType(),
		}},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{{
			Pub:  true,
			Name: "ClientActor",
			Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
		}},
		Data: []*schema.Data{user},
		Services: []*schema.Service{
			{Name: "PublicClientService", Api: true, Audiences: []*schema.ActorAudience{{Actor: "ClientActor"}}, Methods: []*schema.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
			{Name: "InternalClientService", Pub: true, Methods: []*schema.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
		},
	})

	gen := newTestGen(pkg, ".")
	payload := gen.buildServiceTsPayload()
	if len(payload.Services) != 1 {
		t.Fatalf("unexpected service count: %d", len(payload.Services))
	}
	if payload.Services[0].Name != "PublicClientService" {
		t.Fatalf("unexpected service: %s", payload.Services[0].Name)
	}
}

func TestServicesTemplatePassesOptionsDirectly(t *testing.T) {
	payload := &_ServiceTsPayload{
		ExternalTypeImports: []*_TypeImport{{Alias: "userpub", Path: "@acme/skeled-userpub"}},
		Services: []*_Service{{
			SkelName:    "demo.user.UserService",
			SpecName:    "UserServiceSpec",
			FactoryName: "createUserService",
			Methods: []*_ServiceMethod{{
				Name:       "ping",
				SkelName:   "ping",
				HasParams:  false,
				ReturnType: "void",
				ParamDocs: []*_MethodParamDoc{
					{Name: "params", Description: "Must be null"},
					{Name: "options", Description: "Call options, optional"},
				},
			}},
		}},
	}

	output := renderTemplate(t, serviceTsTemplate, payload)
	if !strings.Contains(output, "options,\n      });") {
		t.Fatalf("expected rendered services to pass options directly, got:\n%s", output)
	}
	if !strings.Contains(output, "import type * as userpub from '@acme/skeled-userpub';") {
		t.Fatalf("expected rendered services to import external types, got:\n%s", output)
	}
	for _, check := range []string{
		"import {\n  UserServiceSpec,\n} from './spec';",
		"serviceName: UserServiceSpec.serviceName",
		"methodName: UserServiceSpec.methods.ping",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected rendered services to contain %q, got:\n%s", check, output)
		}
	}
	if strings.Contains(output, "export {};") {
		t.Fatalf("expected rendered services to omit trailing empty export, got:\n%s", output)
	}
}

func TestServicesTemplateInjectsWireOnlyForBinaryMethods(t *testing.T) {
	payload := &_ServiceTsPayload{
		Services: []*_Service{{
			SkelName:    "demo.file.FileService",
			SpecName:    "FileServiceSpec",
			FactoryName: "createFileService",
			Methods: []*_ServiceMethod{
				{
					Name:       "ping",
					SkelName:   "ping",
					ReturnType: "void",
				},
				{
					Name:       "upload",
					SkelName:   "upload",
					HasParams:  true,
					ReturnType: "void",
					HasWire:    true,
					Arguments: []*_MethodArgument{{
						Name: "content",
						Type: &_Type{Plain: "Uint8Array"},
					}},
				},
			},
		}},
	}

	output := renderTemplate(t, serviceTsTemplate, payload)
	for _, check := range []string{
		"methodName: FileServiceSpec.methods.ping,\n        params,\n        options,",
		"methodName: FileServiceSpec.methods.upload,",
		"options: {\n          ...options,\n          wire: FileServiceSpec.wire.upload,\n        },",
	} {
		if !strings.Contains(output, check) {
			t.Fatalf("expected rendered services to contain %q, got:\n%s", check, output)
		}
	}
	if strings.Contains(output, "FileServiceSpec.wire.ping") {
		t.Fatalf("expected JSON method to omit wire, got:\n%s", output)
	}
}
