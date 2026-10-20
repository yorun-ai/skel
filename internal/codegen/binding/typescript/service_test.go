package typescript

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

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
