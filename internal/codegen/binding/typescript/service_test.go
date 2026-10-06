package typescript

import (
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
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
	service := (_Types{}).castService(&model.Service{
		Name:        "UserService",
		SkelName:    "demo.user.UserService",
		Description: "User service",
		Methods: []*model.Method{
			{
				Name:        "getUser",
				Description: "Get a user by ID",
				Arguments: []*model.Argument{
					{
						Name:        "userId",
						Description: "User ID",
						Example:     `"10001"`,
						Type: &model.Type{
							Kind:   model.TypeKindScalar,
							Scalar: model.ScalarInt,
						},
					},
				},
				ResultType: &model.Type{
					Kind: model.TypeKindData,
					Data: &model.Data{
						Name: "User",
					},
					Nullable: true,
				},
				OutputDescription: "User information",
				OutputExample:     `{ id:10001, name:"zhangsan" }`,
				ArgumentsData: &model.Data{
					Name: "UserServiceGetUserArguments",
					Members: []*model.DataMember{
						{
							Name: "userId",
							Type: &model.Type{
								Kind:   model.TypeKindScalar,
								Scalar: model.ScalarInt,
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
	service := (_Types{}).castService(&model.Service{
		Name:             "UserService",
		Deprecated:       true,
		DeprecatedReason: "Use ProfileService instead\nComplete migration first",
		Methods: []*model.Method{{
			Name:             "getUser",
			Deprecated:       true,
			DeprecatedReason: "Use getProfile instead",
			Arguments: []*model.Argument{{
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
	imports := (_Types{}).buildServiceTypeImports([]*model.Service{{
		Methods: []*model.Method{{
			Arguments: []*model.Argument{{
				Type: &model.Type{
					Kind: model.TypeKindData,
					Data: &model.Data{
						Name: "Page",
					},
					TypeArguments: []*model.Type{{
						Kind: model.TypeKindEnum,
						Enum: &model.Enum{Name: "UserStatus"},
					}},
				},
			}},
			ResultType: &model.Type{
				Kind: model.TypeKindList,
				List: &model.ListType{Value: &model.Type{
					Kind: model.TypeKindData,
					Data: &model.Data{
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
	services := []*model.Service{{
		Methods: []*model.Method{{
			Arguments: []*model.Argument{{
				Type: &model.Type{
					Kind:          model.TypeKindData,
					Data:          &model.Data{Name: "UserSummary"},
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
	user := &model.Data{
		Name: "User",
		Members: []*model.DataMember{{
			Name: "id",
			Type: codegentest.IntType(),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{
			{Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
			{Name: "AgentActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaAgent)}},
		},
		Data: []*model.Data{user},
		Services: []*model.Service{
			{Name: "ClientService", Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
			{Name: "AgentService", Audiences: []*model.ActorAudience{{Actor: "AgentActor"}}, Methods: []*model.Method{{
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
	user := &model.Data{
		Pub:  true,
		Name: "User",
		Members: []*model.DataMember{{
			Name: "id",
			Type: codegentest.IntType(),
		}},
	}
	pkg := buildModelDomainForTest(t, model.DomainSpec{
		Name: "demo.user",
		Actors: []*model.Actor{{
			Pub:  true,
			Name: "ClientActor",
			Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)},
		}},
		Data: []*model.Data{user},
		Services: []*model.Service{
			{Name: "PublicClientService", Api: true, Audiences: []*model.ActorAudience{{Actor: "ClientActor"}}, Methods: []*model.Method{{
				Name:       "getUser",
				ResultType: codegentest.DataType(user),
			}}},
			{Name: "InternalClientService", Pub: true, Methods: []*model.Method{{
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
