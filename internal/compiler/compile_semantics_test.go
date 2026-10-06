package compiler

import (
	"path/filepath"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCompileServiceAndData(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": describedUserDomain,
		"user.skel": `
actor ClientActor {
    via client {}
    auth {
        credential {
            subject: string
        }
        info {
            userId: string
        }
    }
}

service UserService {
    for ClientActor via client

    @desc("Get a user by ID")
    method getUserById {
        input {
            userId: string
        }
        output User?
    }
}

data User {
    id: int
    status: UserStatus
}

enum UserStatus {
    PENDING
    ACTIVE
}
`,
	})

	if domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain name: %s", domain.Name())
	}
	if len(domain.Services()) != 1 {
		t.Fatalf("unexpected service count: %d", len(domain.Services()))
	}
	service := domain.Services()[0]
	if service.Name != "UserService" {
		t.Fatalf("unexpected service name: %s", service.Name)
	}
	if len(service.Audiences) != 1 || service.Audiences[0].Actor != "ClientActor" || service.Audiences[0].Via != "client" {
		t.Fatalf("unexpected service audiences: %+v", service.Audiences)
	}
	if len(service.Methods) != 1 {
		t.Fatalf("unexpected method count: %d", len(service.Methods))
	}
	if len(domain.Data()) != 1 {
		t.Fatalf("unexpected data count: %d", len(domain.Data()))
	}
	if len(domain.Enums()) != 1 {
		t.Fatalf("unexpected enum count: %d", len(domain.Enums()))
	}
}

func TestCompileAcrossFilesAndResolveTypes(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": describedUserDomain,
		"types.skel": `
actor ClientActor { via client {} }
actor OpenAPIActor { via openapi {} }

enum UserStatus {
    ACTIVE
}

data Page<TItem> {
    items: list<TItem>
    nextToken: string?
}

data User {
    id: int
    status: UserStatus
}
`,
		"service.skel": `
service UserService {
    for ClientActor
    for OpenAPIActor

    @desc("List users")
    method listUsers {
        @desc("Query criteria")
        input {
            pageToken: string
        }
        @desc("Paginated result")
        output Page<User>
    }
}
`,
	})

	if len(domain.Data()) != 2 {
		t.Fatalf("unexpected data count: %d", len(domain.Data()))
	}
	if len(domain.Services()) != 1 {
		t.Fatalf("unexpected service count: %d", len(domain.Services()))
	}

	service := domain.Services()[0]
	method := service.Methods[0]
	if service.SkelName != "demo.user.UserService" {
		t.Fatalf("unexpected skel name: %q", service.SkelName)
	}
	if len(service.Audiences) != 2 {
		t.Fatalf("unexpected for count: %d", len(service.Audiences))
	}
	if service.Audiences[0].Actor != "ClientActor" || service.Audiences[1].Actor != "OpenAPIActor" {
		t.Fatalf("unexpected service audiences: %+v", service.Audiences)
	}
	if method.Description != "List users" {
		t.Fatalf("unexpected method description: %q", method.Description)
	}
	if method.InputDescription != "Query criteria" {
		t.Fatalf("unexpected input description: %q", method.InputDescription)
	}
	if method.OutputDescription != "Paginated result" {
		t.Fatalf("unexpected output description: %q", method.OutputDescription)
	}
	if method.ArgumentsData == nil {
		t.Fatal("arguments data should not be nil")
	}
	if method.ArgumentsData.Name != "UserServiceListUsersArguments" {
		t.Fatalf("unexpected arguments data name: %q", method.ArgumentsData.Name)
	}

	result := method.ResultType
	if result.Kind != schema.TypeKindData {
		t.Fatalf("unexpected result kind: %v", result.Kind)
	}
	if result.Data == nil || result.Data.Name != "Page" {
		t.Fatalf("unexpected result data: %+v", result.Data)
	}
	if len(result.TypeArguments) != 1 {
		t.Fatalf("unexpected type arg count: %d", len(result.TypeArguments))
	}
	if result.TypeArguments[0].Kind != schema.TypeKindData || result.TypeArguments[0].Data.Name != "User" {
		t.Fatalf("unexpected type argument: %+v", result.TypeArguments[0])
	}
	if result.Name() != "PageOfUser" {
		t.Fatalf("unexpected result type name: %q", result.Name())
	}

	pageData := findDataByName(t, domain, "Page")
	if len(pageData.TypeParameters) != 1 || pageData.TypeParameters[0].Name != "TItem" {
		t.Fatalf("unexpected type parameters: %+v", pageData.TypeParameters)
	}
	if pageData.Members[0].Type.Kind != schema.TypeKindList {
		t.Fatalf("unexpected items member kind: %v", pageData.Members[0].Type.Kind)
	}
	if pageData.Members[0].Type.List.Value.Kind != schema.TypeKindTypeParameter {
		t.Fatalf("unexpected list value kind: %v", pageData.Members[0].Type.List.Value.Kind)
	}
	if pageData.Members[0].Type.List.Value.TypeParameter.Name != "TItem" {
		t.Fatalf("unexpected type parameter name: %q", pageData.Members[0].Type.List.Value.TypeParameter.Name)
	}
	if !pageData.Members[1].Type.Nullable {
		t.Fatal("nextToken should be nullable")
	}

	userData := findDataByName(t, domain, "User")
	if userData.Members[1].Type.Kind != schema.TypeKindEnum {
		t.Fatalf("unexpected user status kind: %v", userData.Members[1].Type.Kind)
	}
	if userData.Members[1].Type.Enum.Name != "UserStatus" {
		t.Fatalf("unexpected enum name: %q", userData.Members[1].Type.Enum.Name)
	}
}

func TestCompileDirectoryIncludesAllDeclarationKinds(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "domain.skel"), describedUserDomain)
	writeFile(t, filepath.Join(dir, "service.skel"), "domain demo.user\nactor PortalAdminActor { via client {} }\nservice AgentService { for PortalAdminActor\nmethod ping {} }\n")
	writeFile(t, filepath.Join(dir, "types.skel"), "domain demo.user\ndata User { id: int }\nenum UserStatus { ACTIVE }\n")

	result, err := Compile(Option{SkelIn: dir})
	if err != nil {
		t.Fatalf("compile directory: %v", err)
	}
	domain := result.Domain

	entryKinds := map[string]bool{}
	for _, service := range domain.Services() {
		entryKinds["service"] = true
		if service.Name != "AgentService" {
			t.Fatalf("unexpected service: %+v", service)
		}
	}
	for _, actor := range domain.Actors() {
		entryKinds["actor"] = true
		if actor.Name != "PortalAdminActor" {
			t.Fatalf("unexpected actor: %+v", actor)
		}
	}
	for _, dataType := range domain.Data() {
		if dataType.Name == "User" {
			entryKinds["data"] = true
		}
	}
	for _, enumType := range domain.Enums() {
		if enumType.Name == "UserStatus" {
			entryKinds["enum"] = true
		}
	}
	if !entryKinds["service"] || !entryKinds["data"] || !entryKinds["enum"] || !entryKinds["actor"] {
		t.Fatalf("unexpected parsed entry kinds: %+v", entryKinds)
	}
}
