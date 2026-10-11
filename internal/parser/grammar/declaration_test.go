package grammar

import (
	"strings"
	"testing"
)

func TestParseSkelContent(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

@desc("User status")
pub enum UserStatus {
    @desc("Active")
    ACTIVE
    DISABLED
}

@desc("Pagination information")
pub data Page<TItem> {
    @desc("Data item")
    items: list<TItem>
    nextToken: string?
}

@desc("Database configuration")
pub config DatabaseConfig eternal {
    @desc("Connection address")
    connUrl: string
}

@desc("User created event")
	pub event UserCreatedEvent {
    @sensitive
    payload {
        @desc("User ID")
        userId: int
    }
}

pub actor PortalAdminActor {
    auth {
        @sensitive
        credential {
            subject: string
        }
        @sensitive
        info {
            @desc("User ID")
            userId: int
        }
    }
}

pub service UserService {
    for PortalAdminActor, LinkActor

    @desc("List users")
    method listUsers {
        @desc("Query criteria")
        input {
            @desc("Pagination cursor")
            @example("cursor-1")
            pageToken: string?
        }
        @desc("Paginated result")
        output map<string, Page<User>>
    }
}

task RebuildUserIndexTask {
    @desc("Scheduled trigger")
    trigger atTime {
        input {
            startAt: localdatetime
        }
    }
}

web UserPortalWeb {
    for PortalAdminActor, LinkActor
}
`)

	if len(content.Entries) != 8 {
		t.Fatalf("unexpected entry count: %d", len(content.Entries))
	}

	enumEntry := content.Entries[0]
	if enumEntry.Enum == nil || enumEntry.Enum.Name.Value != "UserStatus" {
		t.Fatalf("unexpected enum entry: %+v", enumEntry)
	}
	if !enumEntry.Enum.Pub {
		t.Fatal("expected pub enum")
	}
	if len(enumEntry.Enum.Decorators) != 1 || enumEntry.Enum.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected enum decorators: %+v", enumEntry.Enum.Decorators)
	}
	if len(enumEntry.Enum.Items) != 2 || enumEntry.Enum.Items[0].Name.Value != "ACTIVE" {
		t.Fatalf("unexpected enum items: %+v", enumEntry.Enum.Items)
	}
	if len(enumEntry.Enum.Items[0].Decorators) != 1 || enumEntry.Enum.Items[0].Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected enum item decorators: %+v", enumEntry.Enum.Items[0].Decorators)
	}

	dataEntry := content.Entries[1]
	if dataEntry.Data == nil || dataEntry.Data.Name.Value != "Page" {
		t.Fatalf("unexpected data entry: %+v", dataEntry)
	}
	if !dataEntry.Data.Pub {
		t.Fatal("expected pub data")
	}
	if len(dataEntry.Data.Decorators) != 1 || dataEntry.Data.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected data decorators: %+v", dataEntry.Data.Decorators)
	}
	if len(dataEntry.Data.TypeParameters) != 1 || dataEntry.Data.TypeParameters[0].Name.Value != "TItem" {
		t.Fatalf("unexpected type parameters: %+v", dataEntry.Data.TypeParameters)
	}
	if dataEntry.Data.Members[0].Type.List == nil || dataEntry.Data.Members[0].Type.List.Value.Reference.Name.String() != "TItem" {
		t.Fatalf("unexpected data member type: %+v", dataEntry.Data.Members[0].Type)
	}
	if len(dataEntry.Data.Members[0].Decorators) != 1 || dataEntry.Data.Members[0].Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected data member decorators: %+v", dataEntry.Data.Members[0].Decorators)
	}
	if !dataEntry.Data.Members[1].Type.Nullable {
		t.Fatal("expected nullable nextToken member")
	}

	configEntry := content.Entries[2]
	if configEntry.Config == nil || configEntry.Config.Name.Value != "DatabaseConfig" {
		t.Fatalf("unexpected config entry: %+v", configEntry)
	}
	if !configEntry.Config.Pub {
		t.Fatal("expected pub config")
	}
	if len(configEntry.Config.Decorators) != 1 || configEntry.Config.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected config decorators: %+v", configEntry.Config.Decorators)
	}
	if len(configEntry.Config.TypeParameters) != 0 {
		t.Fatalf("unexpected config type parameters: %+v", configEntry.Config.TypeParameters)
	}
	if configEntry.Config.Qualifier == nil || configEntry.Config.Qualifier.Value != "eternal" {
		t.Fatalf("unexpected config qualifier: %+v", configEntry.Config.Qualifier)
	}

	eventEntry := content.Entries[3]
	if eventEntry.Event == nil || eventEntry.Event.Name.Value != "UserCreatedEvent" {
		t.Fatalf("unexpected event entry: %+v", eventEntry)
	}
	if len(eventEntry.Event.Decorators) != 1 || eventEntry.Event.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected event decorators: %+v", eventEntry.Event.Decorators)
	}
	if !eventEntry.Event.Pub {
		t.Fatal("expected pub event")
	}
	if eventEntry.Event.Payload == nil || len(eventEntry.Event.Payload.Members) != 1 || eventEntry.Event.Payload.Members[0].Name.Value != "userId" {
		t.Fatalf("unexpected event payload: %+v", eventEntry.Event.Payload)
	}
	if len(eventEntry.Event.Payload.Decorators) != 1 || eventEntry.Event.Payload.Decorators[0].Name.Value != "sensitive" {
		t.Fatalf("unexpected event payload decorators: %+v", eventEntry.Event.Payload.Decorators)
	}

	actorEntry := content.Entries[4]
	if actorEntry.Actor == nil || actorEntry.Actor.Name.Value != "PortalAdminActor" {
		t.Fatalf("unexpected actor entry: %+v", actorEntry)
	}
	if !actorEntry.Actor.Pub {
		t.Fatal("expected pub actor")
	}
	if len(actorEntry.Actor.Sections) != 1 {
		t.Fatalf("unexpected actor sections: %+v", actorEntry.Actor.Sections)
	}
	auth := actorEntry.Actor.Sections[0].Auth
	if auth == nil {
		t.Fatalf("unexpected actor auth: %+v", actorEntry.Actor.Sections[0])
	}
	if len(auth.Credential.Decorators) != 1 || auth.Credential.Decorators[0].Name.Value != "sensitive" {
		t.Fatalf("unexpected actor credential decorators: %+v", auth.Credential.Decorators)
	}
	if len(auth.Info.Decorators) != 1 || auth.Info.Decorators[0].Name.Value != "sensitive" {
		t.Fatalf("unexpected actor info decorators: %+v", auth.Info.Decorators)
	}
	if auth.Credential == nil || len(auth.Credential.Members) != 1 || auth.Credential.Members[0].Name.Value != "subject" {
		t.Fatalf("unexpected actor credential: %+v", actorEntry.Actor.Sections[0])
	}
	if auth.Info == nil || len(auth.Info.Members) != 1 || auth.Info.Members[0].Name.Value != "userId" {
		t.Fatalf("unexpected actor info: %+v", auth)
	}
	if len(auth.Info.Members[0].Decorators) != 1 || auth.Info.Members[0].Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected actor info decorators: %+v", auth.Info.Members[0].Decorators)
	}

	serviceEntry := content.Entries[5]
	if serviceEntry.Service == nil || serviceEntry.Service.Name.Value != "UserService" {
		t.Fatalf("unexpected service entry: %+v", serviceEntry)
	}
	if !serviceEntry.Service.Pub {
		t.Fatal("expected pub service")
	}
	audiences := serviceAudiencesForTest(serviceEntry.Service)
	if len(audiences) != 1 || len(audiences[0].Actors) != 2 {
		t.Fatalf("unexpected actor count: %d", len(audiences))
	}
	if audiences[0].Actors[0].Actor.String() != "PortalAdminActor" || audiences[0].Actors[1].Actor.String() != "LinkActor" {
		t.Fatalf("unexpected actors: %+v", audiences)
	}
	methods := serviceMethodsForTest(serviceEntry.Service)
	if len(methods) != 1 {
		t.Fatalf("unexpected method count: %d", len(methods))
	}

	method := methods[0]
	if method.Name.Value != "listUsers" {
		t.Fatalf("unexpected method name: %s", method.Name.Value)
	}
	if len(method.Decorators) != 1 || method.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected method decorators: %+v", method.Decorators)
	}
	input := methodInputForTest(method)
	if input == nil || len(input.Arguments) != 1 {
		t.Fatalf("unexpected input section: %+v", input)
	}
	if input.Arguments[0].Name.Value != "pageToken" {
		t.Fatalf("unexpected method arg: %+v", input.Arguments[0])
	}
	if !input.Arguments[0].Type.Nullable {
		t.Fatal("expected nullable pageToken arg")
	}
	if len(input.Decorators) != 1 || input.Decorators[0].Name.Value != "desc" {
		t.Fatalf("unexpected input decorators: %+v", input.Decorators)
	}
	if len(input.Arguments[0].Decorators) != 2 {
		t.Fatalf("unexpected argument decorators: %+v", input.Arguments[0].Decorators)
	}
	output := methodOutputForTest(method)
	if output == nil || output.Type.Map == nil {
		t.Fatalf("expected output map type: %+v", output)
	}
	if output.Type.Map.Key.Plain == nil || *output.Type.Map.Key.Plain != String {
		t.Fatalf("unexpected map key type: %+v", output.Type.Map.Key)
	}
	if output.Type.Map.Value.Reference == nil || output.Type.Map.Value.Reference.Name.String() != "Page" {
		t.Fatalf("unexpected map value type: %+v", output.Type.Map.Value)
	}
	if len(output.Type.Map.Value.Reference.TypeArguments) != 1 || output.Type.Map.Value.Reference.TypeArguments[0].Reference.Name.String() != "User" {
		t.Fatalf("unexpected type arguments: %+v", output.Type.Map.Value.Reference.TypeArguments)
	}

	taskEntry := content.Entries[6]
	if taskEntry.Task == nil || taskEntry.Task.Name.Value != "RebuildUserIndexTask" {
		t.Fatalf("unexpected task entry: %+v", taskEntry)
	}
	if len(taskEntry.Task.Triggers) != 1 || taskEntry.Task.Triggers[0].Name.Value != "atTime" {
		t.Fatalf("unexpected task triggers: %+v", taskEntry.Task.Triggers)
	}
	if taskEntry.Task.Triggers[0].Input == nil || len(taskEntry.Task.Triggers[0].Input.Arguments) != 1 {
		t.Fatalf("unexpected task input: %+v", taskEntry.Task.Triggers[0].Input)
	}

	webEntry := content.Entries[7]
	if webEntry.Web == nil || webEntry.Web.Name.Value != "UserPortalWeb" {
		t.Fatalf("unexpected web entry: %+v", webEntry)
	}
	if len(webEntry.Web.Audiences) != 1 || len(webEntry.Web.Audiences[0].Actors) != 2 || webEntry.Web.Audiences[0].Actors[0].Actor.String() != "PortalAdminActor" || webEntry.Web.Audiences[0].Actors[1].Actor.String() != "LinkActor" {
		t.Fatalf("unexpected web audiences: %+v", webEntry.Web.Audiences)
	}
}

// TODO: Remove this compatibility test when legacy via syntax and warnings are removed.
func TestParseActorAcceptsLegacyViaSections(t *testing.T) {
	content := parseSkelForTest(t, `domain demo.user
actor ClientActor {
    via client {}
    via openapi {}
    auth { credential { subject: string } }
    via Custom {}
}`)
	sections := content.Entries[0].Actor.Sections
	if len(sections) != 4 || sections[2].Auth == nil {
		t.Fatalf("unexpected actor sections: %+v", sections)
	}
	for _, test := range []struct {
		index int
		name  string
	}{{0, "client"}, {1, "openapi"}, {3, "Custom"}} {
		via := sections[test.index].Via
		if via == nil || via.Name.Value != test.name {
			t.Fatalf("unexpected legacy via section %d: %+v", test.index, via)
		}
	}
}

func TestParseActorAuthCredentialAndInfoCanAppearInEitherOrder(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

actor ClientActor {
    auth {
        info {
            userId: int
        }
        credential {
            subject: string
        }
    }
}
	`)
	actor := content.Entries[0].Actor
	if len(actor.Sections) != 1 || actor.Sections[0].Auth == nil || actor.Sections[0].Auth.Info == nil || actor.Sections[0].Auth.Credential == nil {
		t.Fatalf("unexpected actor sections: %+v", actor.Sections)
	}
}

func TestParseActorPermissionMustBeEmpty(t *testing.T) {
	parser := buildSkelParserForTest(t)
	_, err := parser.ParseString("test.skel", strings.TrimSpace(`
domain demo.user

actor ClientActor {
    permission {
        User:read
    }
}
`))
	if err == nil {
		t.Fatal("expected parse error when actor permission block is not empty")
	}
}

func TestParseResourceCheckUsesInputBlock(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

resource User {
    @desc("Lookup")
    check byId {
        @desc("Lookup input")
        input {
            @sensitive
            id: int
        }
    }
    action read
}
`)
	check := content.Entries[0].Resource.Checks[0]
	if check.Name.Value != "byId" || check.Input == nil || len(check.Input.Arguments) != 1 {
		t.Fatalf("unexpected resource check: %+v", check)
	}
	if len(check.Decorators) != 1 || len(check.Input.Decorators) != 1 || len(check.Input.Arguments[0].Decorators) != 1 {
		t.Fatalf("resource check decorators were not preserved: %+v", check)
	}

}

func serviceAudiencesForTest(service *Service) []*ActorAudience {
	audiences := make([]*ActorAudience, 0)
	for _, section := range service.Sections {
		if section.Audience != nil {
			audiences = append(audiences, section.Audience)
		}
	}
	return audiences
}

func serviceMethodsForTest(service *Service) []*Method {
	methods := make([]*Method, 0)
	for _, section := range service.Sections {
		if section.Method != nil {
			section.Method.Decorators = append(section.Decorators, section.Method.Decorators...)
			methods = append(methods, section.Method)
		}
	}
	return methods
}

func methodInputForTest(method *Method) *MethodInput {
	return method.Input
}

func methodOutputForTest(method *Method) *MethodOutput {
	return method.Output
}

func TestWebMountSyntax(t *testing.T) {
	for _, body := range []string{
		"for ClientActor\nmount /portal/v1-assets/",
		"for ClientActor\nmount /* prefix */ /portal/v1-assets/ // trailing comment",
		"for ClientActor mount /portal/v1-assets/",
	} {
		content := parseSkelForTest(t, "domain demo\nweb PortalWeb {\n"+body+"\n}")
		web := content.Entries[0].Web
		if len(web.Mounts) != 1 || web.Mounts[0].Path.Value != "/portal/v1-assets/" || len(web.Audiences) != 1 {
			t.Fatalf("unexpected web: %+v", web)
		}
	}
	content := parseSkelForTest(t, "domain demo\nweb PortalWeb {\n    mount /\n    for ClientActor\n}")
	path := content.Entries[0].Web.Mounts[0].Path
	if path.Value != "/" || path.Pos.Filename != "test.skel" || path.Pos.Line != 3 || path.Pos.Column != 11 {
		t.Fatalf("unexpected path or position: %+v", path)
	}
}
