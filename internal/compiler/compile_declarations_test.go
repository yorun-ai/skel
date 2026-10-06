package compiler

import (
	"go.yorun.ai/skel/schema"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileServiceWithoutActor(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": describedUserDomain,
		"service.skel": `
service UserService {
    method ping {
        output string
    }
}
`,
	})

	if len(domain.Services()) != 1 {
		t.Fatalf("unexpected service count: %d", len(domain.Services()))
	}
	service := domain.Services()[0]
	if service.Name != "UserService" {
		t.Fatalf("unexpected service name: %s", service.Name)
	}
	if len(service.Audiences) != 0 {
		t.Fatalf("expected empty audiences, got %+v", service.Audiences)
	}
}

func TestCompileEvent(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": describedUserDomain,
		"event.skel": `
actor PartnerActor { via client {} }
actor OpenAPIActor { via openapi {} }

@desc("User created event")
event UserCreatedEvent {
    payload {
        @desc("User ID")
        userId: int

        @desc("Created at")
        createdAt: localdatetime
    }
}
`,
	})

	if len(domain.Events()) != 1 {
		t.Fatalf("unexpected event count: %d", len(domain.Events()))
	}
	event := domain.Events()[0]
	if event.Name != "UserCreatedEvent" {
		t.Fatalf("unexpected event: %+v", event)
	}
	if event.Kind != schema.DataKindEvent {
		t.Fatalf("unexpected event kind: %v", event.Kind)
	}
	if len(event.Members) != 2 {
		t.Fatalf("unexpected event members: %+v", event.Members)
	}
}

func TestCompileTask(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": describedUserDomain,
		"task.skel": `
@desc("Rebuild the user index")
task RebuildUserIndexTask {
	@desc("Scheduled trigger")
	trigger atTime {
		input {
			startAt: localdatetime
		}
	}

	@desc("Trigger by user group")
	trigger forGroup {
		input {
			groupId: int
		}
	}
}

`,
	})

	if len(domain.Tasks()) != 1 {
		t.Fatalf("unexpected task count: %d", len(domain.Tasks()))
	}
	task := domain.Tasks()[0]
	if task.Name != "RebuildUserIndexTask" || task.SkelName != "demo.user.RebuildUserIndexTask" {
		t.Fatalf("unexpected task: %+v", task)
	}
	if len(task.Triggers) != 2 {
		t.Fatalf("unexpected task trigger count: %d", len(task.Triggers))
	}
	if task.Triggers[0].Name != "atTime" || task.Triggers[1].Name != "forGroup" {
		t.Fatalf("unexpected task triggers: %+v", task.Triggers)
	}
}

func TestCompileDeprecatedMetadata(t *testing.T) {
	domain := parseDomain(t, map[string]string{
		"domain.skel": "domain demo\n",
		"deprecated.skel": `
domain demo

@deprecated("Use NewStatus")
enum Status {
    @deprecated("Use ACTIVE")
    LEGACY
}

@deprecated("Use Profile")
data User {
    @deprecated("Use id")
    legacyId: string
}

@deprecated("Use NewConfig")
config AppConfig eternal {}

@deprecated("Use NewEvent")
event ChangedEvent {
    payload {}
}

@deprecated("Use NewActor")
actor ClientActor {
    via client {}
}

@deprecated("Use NewResource")
resource UserResource {
    @deprecated("Use read")
    action legacyRead

    @deprecated("Use byId")
    check byLegacyId {
        input {
            @deprecated("Use id")
            legacyId: string
        }
    }
}

@deprecated("Use ProfileService")
service UserService {
    @deprecated("Use getProfile")
    method getUser {
        input {
            @deprecated("Use id")
            legacyId: string
        }
    }
}

@deprecated("Use NewTask")
task RefreshTask {
    @deprecated("Use manually")
    trigger legacy {
        input {
            @deprecated("Use force")
            legacyForce: bool
        }
    }
}

@deprecated("Use NewWeb")
web PortalWeb {
    for ClientActor via client
}
`,
	})

	if enum := domain.Enums()[0]; !enum.Deprecated || !enum.Items[0].Deprecated {
		t.Fatalf("unexpected enum deprecation: %+v", enum)
	}
	if data := domain.Data()[0]; !data.Deprecated || !data.Members[0].Deprecated {
		t.Fatalf("unexpected data deprecation: %+v", data)
	}
	if !domain.Configs()[0].Deprecated || !domain.Events()[0].Deprecated || !domain.Actors()[0].Deprecated {
		t.Fatal("expected config, event, and actor deprecation")
	}
	resource := domain.Resources()[0]
	if !resource.Deprecated || !resource.Actions[0].Deprecated || !resource.Checks[0].Deprecated ||
		!resource.Checks[0].Method.Arguments[1].Deprecated {
		t.Fatalf("unexpected resource deprecation: resource=%t action=%t check=%t argument=%t",
			resource.Deprecated, resource.Actions[0].Deprecated, resource.Checks[0].Deprecated,
			resource.Checks[0].Method.Arguments[1].Deprecated)
	}
	service := domain.Services()[0]
	if !service.Deprecated || !service.Methods[0].Deprecated || !service.Methods[0].Arguments[0].Deprecated {
		t.Fatalf("unexpected service deprecation: %+v", service)
	}
	task := domain.Tasks()[0]
	if !task.Deprecated || !task.Triggers[0].Deprecated || !task.Triggers[0].Arguments[0].Deprecated {
		t.Fatalf("unexpected task deprecation: %+v", task)
	}
	if !domain.Webs()[0].Deprecated {
		t.Fatalf("unexpected web deprecation: %+v", domain.Webs()[0])
	}
}

func TestCompileRejectsDeprecatedOnStructuralBlocks(t *testing.T) {
	tests := map[string]string{
		"domain": `@deprecated("Not supported")
domain demo
`,
		"input": `domain demo
service UserService {
    method getUser {
        @deprecated("Not supported")
        input { id: string }
    }
}
`,
		"output": `domain demo
service UserService {
    method getUser {
        @deprecated("Not supported")
        output string
    }
}
`,
		"payload": `domain demo
event ChangedEvent {
    @deprecated("Not supported")
    payload { id: string }
}
`,
		"credential": `domain demo
actor ClientActor {
    via client {}
    auth {
        @deprecated("Not supported")
        credential { token: string }
        info { id: string }
    }
}
`,
		"info": `domain demo
actor ClientActor {
    via client {}
    auth {
        credential { token: string }
        @deprecated("Not supported")
        info { id: string }
    }
}
`,
	}

	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if name == "domain" {
				writeFile(t, filepath.Join(dir, "domain.skel"), source)
			} else {
				writeFile(t, filepath.Join(dir, "domain.skel"), "domain demo\n")
				writeFile(t, filepath.Join(dir, "contract.skel"), source)
			}
			_, err := Compile(Option{SkelIn: dir})
			if err == nil || !strings.Contains(err.Error(), "unexpected decorator @deprecated") {
				t.Fatalf("expected unsupported @deprecated diagnostic, got %v", err)
			}
		})
	}
}
