package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func TestAnalyze(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Enum: &grammar.Enum{
					Name: ident("UserStatus"),
					Items: []*grammar.EnumItem{
						{Name: ident("ACTIVE")},
					},
				},
			},
			{
				Data: &grammar.Data{
					Name: ident("User"),
					Members: []*grammar.DataMember{
						{Name: ident("id"), Type: plainType(grammar.Int)},
						{Name: ident("status"), Type: refGrammarType("UserStatus")},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("UserConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("displayName"), Type: plainType(grammar.String)},
					},
				},
			},
			{
				Actor: &grammar.Actor{
					Name: ident("PortalAdminActor"),
					Vias: []*grammar.ActorVia{
						{Name: ident("client")},
					},
					Sections: []*grammar.ActorSection{
						grammarActorAuthSection(
							[]*grammar.DataMember{{Name: ident("subject"), Type: plainType(grammar.String)}},
							[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
						),
					},
				},
			},
			{
				Event: &grammar.Event{
					Name: ident("UserCreatedEvent"),
					Payload: &grammar.EventPayload{
						Members: []*grammar.DataMember{
							{Name: ident("userId"), Type: plainType(grammar.Int)},
						},
					},
				},
			},
			{
				Service: &grammar.Service{
					Auth:      &grammar.AuthMarker{Value: "required"},
					Api:       true,
					Name:      ident("UserApiService"),
					Audiences: []*grammar.ServiceAudience{serviceAllow("PortalAdminActor")},
					Methods: []*grammar.Method{
						{
							Name: ident("getUser"),
							Input: &grammar.MethodInput{
								Arguments: []*grammar.Argument{
									{Name: ident("userId"), Type: plainType(grammar.Int)},
								},
							},
							Output: &grammar.MethodOutput{
								Type: refGrammarType("User"),
							},
						},
					},
				},
			},
			{
				Task: &grammar.Task{
					Name: ident("RebuildUserIndexTask"),
					Triggers: []*grammar.TaskTrigger{
						{
							Name: ident("atTime"),
							Input: &grammar.MethodInput{
								Arguments: []*grammar.Argument{
									{Name: ident("startAt"), Type: plainType(grammar.LocalDateTime)},
								},
							},
						},
					},
				},
			},
		},
	}).Schema()

	if domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain name: %s", domain.Name())
	}
	if domain.Description() != "" {
		t.Fatalf("unexpected domain description: %q", domain.Description())
	}
	if len(domain.Enums()) != 1 || domain.Enums()[0].Name != "UserStatus" {
		t.Fatalf("unexpected enums: %+v", domain.Enums())
	}
	if domain.Enums()[0].SkelName != "demo.user.UserStatus" {
		t.Fatalf("unexpected enum skel name: %q", domain.Enums()[0].SkelName)
	}
	if len(domain.Data()) != 1 || domain.Data()[0].Name != "User" {
		t.Fatalf("unexpected data: %+v", domain.Data())
	}
	if domain.Data()[0].SkelName != "demo.user.User" {
		t.Fatalf("unexpected data skel name: %q", domain.Data()[0].SkelName)
	}
	if domain.Data()[0].Members[1].Type.SkelName != "demo.user.UserStatus" {
		t.Fatalf("unexpected enum ref skel name: %q", domain.Data()[0].Members[1].Type.SkelName)
	}
	if len(domain.Configs()) != 1 || domain.Configs()[0].Name != "UserConfig" {
		t.Fatalf("unexpected configs: %+v", domain.Configs())
	}
	if domain.Configs()[0].SkelName != "demo.user.UserConfig" {
		t.Fatalf("unexpected config skel name: %q", domain.Configs()[0].SkelName)
	}
	if len(domain.Events()) != 1 || domain.Events()[0].Name != "UserCreatedEvent" {
		t.Fatalf("unexpected events: %+v", domain.Events())
	}
	if domain.Events()[0].SkelName != "demo.user.UserCreatedEvent" {
		t.Fatalf("unexpected event skel name: %q", domain.Events()[0].SkelName)
	}
	if len(domain.Actors()) != 1 || domain.Actors()[0].Name != "PortalAdminActor" {
		t.Fatalf("unexpected actors: %+v", domain.Actors())
	}
	if domain.Actors()[0].SkelName != "demo.user.PortalAdminActor" {
		t.Fatalf("unexpected actor skel name: %q", domain.Actors()[0].SkelName)
	}
	if domain.Actors()[0].Auth.Credential == nil || domain.Actors()[0].Auth.Credential.SkelName != "demo.user.PortalAdminActorCredential" {
		t.Fatalf("unexpected actor credential: %+v", domain.Actors()[0].Auth.Credential)
	}
	if domain.Actors()[0].Auth.Info == nil || domain.Actors()[0].Auth.Info.SkelName != "demo.user.PortalAdminActorInfo" {
		t.Fatalf("unexpected actor info: %+v", domain.Actors()[0].Auth.Info)
	}
	if len(domain.Services()) != 1 || domain.Services()[0].SkelName != "demo.user.UserApiService" {
		t.Fatalf("unexpected services: %+v", domain.Services())
	}
	authService := domain.Actors()[0].Auth.Service
	if authService.Name != "PortalAdminActorAuthService" || authService.SkelName != "demo.user.PortalAdminActorAuthService" {
		t.Fatalf("unexpected actor auth service: %+v", authService)
	}
	if len(authService.Methods) != 1 || authService.Methods[0].Name != "auth" || authService.Methods[0].SkelName != "auth" {
		t.Fatalf("unexpected actor auth service methods: %+v", authService.Methods)
	}
	credentialMethod := authService.Methods[0]
	if domain.Actors()[0].Auth.Method != credentialMethod {
		t.Fatalf("unexpected actor auth method: %+v", domain.Actors()[0].Auth.Method)
	}
	if len(credentialMethod.Arguments) != 1 || credentialMethod.Arguments[0].Name != "credential" {
		t.Fatalf("unexpected actor credential method arguments: %+v", credentialMethod.Arguments)
	}
	if credentialMethod.Arguments[0].Type.Kind != schema.TypeKindData || credentialMethod.Arguments[0].Type.Data != domain.Actors()[0].Auth.Credential {
		t.Fatalf("unexpected actor credential method argument type: %+v", credentialMethod.Arguments[0].Type)
	}
	if credentialMethod.ResultType.Kind != schema.TypeKindData || credentialMethod.ResultType.Data != domain.Actors()[0].Auth.Info {
		t.Fatalf("unexpected actor credential method result type: %+v", credentialMethod.ResultType)
	}
	if credentialMethod.ArgumentsData == nil || credentialMethod.ArgumentsData.Name != "PortalAdminActorAuthServiceAuthArguments" {
		t.Fatalf("unexpected actor credential method arguments data: %+v", credentialMethod.ArgumentsData)
	}
	if domain.Services()[0].Methods[0].ResultType.Kind != schema.TypeKindData {
		t.Fatalf("unexpected service result type: %+v", domain.Services()[0].Methods[0].ResultType)
	}
	if domain.Services()[0].Methods[0].ResultType.SkelName != "demo.user.User" {
		t.Fatalf("unexpected result type skel name: %q", domain.Services()[0].Methods[0].ResultType.SkelName)
	}
	if len(domain.Tasks()) != 1 || domain.Tasks()[0].SkelName != "demo.user.RebuildUserIndexTask" {
		t.Fatalf("unexpected tasks: %+v", domain.Tasks())
	}
	if len(domain.Tasks()[0].Triggers) != 1 || domain.Tasks()[0].Triggers[0].Arguments[0].Type.Kind != schema.TypeKindScalar {
		t.Fatalf("unexpected task triggers: %+v", domain.Tasks()[0].Triggers)
	}
}

func TestAnalyzeKeepsDomainDescription(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContentWithDescription("demo.user", "User domain"),
	}).Schema()

	if domain.Description() != "User domain" {
		t.Fatalf("unexpected domain description: %q", domain.Description())
	}
}
