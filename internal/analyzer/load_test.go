package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestAnalyzeReturnsErrorWhenEventDoesNotEndWithEvent(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "Event name must end with Event", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Event: &grammar.Event{
					Name: ident("UserCreated"),
					Payload: &grammar.EventPayload{
						Members: []*grammar.DataMember{
							{Name: ident("userId"), Type: plainType(grammar.Int)},
						},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorWhenEventHasTypeParameters(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "does not support type parameters", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Event: &grammar.Event{
					Name:           ident("UserCreatedEvent"),
					TypeParameters: []*grammar.TypeParameter{{Name: ident("TItem")}},
					Payload: &grammar.EventPayload{
						Members: []*grammar.DataMember{
							{Name: ident("item"), Type: refGrammarType("TItem")},
						},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorWhenEventHasQualifier(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "does not support qualifier", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Event: &grammar.Event{
					Name:      ident("UserCreatedEvent"),
					Qualifier: ident("eternal"),
					Payload: &grammar.EventPayload{
						Members: []*grammar.DataMember{
							{Name: ident("userId"), Type: plainType(grammar.Int)},
						},
					},
				},
			},
		},
	})
}

func TestAnalyzeRejectsPubTask(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "task RebuildUserIndexTask does not support pub", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Pub: true,
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
	})
}

func TestAnalyzeReturnsErrorWhenAllowViaDoesNotExist(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, `references undefined actor via "openapi"`, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
					Vias: []*grammar.ActorVia{
						{Name: ident("client")},
					},
				},
			},
			{
				Service: &grammar.Service{
					Name:      ident("UserService"),
					Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor", "openapi")},
					Methods: []*grammar.Method{
						{Name: ident("ping")},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorWhenActorAuthServiceNameConflicts(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, `duplicated identifier "ClientActorAuthService"`, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
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
				Service: &grammar.Service{
					Name: ident("ClientActorAuthService"),
					Methods: []*grammar.Method{
						{Name: ident("ping")},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorWhenActorCredentialNameConflicts(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, `duplicated identifier "ClientActorCredential"`, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: actorWithCredentialForTest("ClientActor"),
			},
			{
				Data: &grammar.Data{
					Name: ident("ClientActorCredential"),
					Members: []*grammar.DataMember{
						{Name: ident("token"), Type: plainType(grammar.String)},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorWhenActorInfoNameConflicts(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, `duplicated identifier "ClientActorInfo"`, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: actorWithCredentialForTest("ClientActor"),
			},
			{
				Data: &grammar.Data{
					Name: ident("ClientActorInfo"),
					Members: []*grammar.DataMember{
						{Name: ident("userId"), Type: plainType(grammar.Int)},
					},
				},
			},
		},
	})
}

func TestAnalyzeReturnsErrorForDuplicatedIdentifier(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, `duplicated identifier "User"`, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Enum: &grammar.Enum{
					Name: ident("User"),
					Items: []*grammar.EnumItem{
						{Name: ident("ACTIVE")},
					},
				},
			},
			{
				Data: &grammar.Data{
					Name: ident("User"),
				},
			},
		},
	})
}
