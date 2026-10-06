package hasher

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/analyzer"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func newHashTestDomain(t *testing.T, serviceDescription string) *model.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Data: &grammar.Data{
					Name: ident("UserProfile"),
					Members: []*grammar.DataMember{
						{Name: ident("userId"), Type: plainType(grammar.String)},
					},
				},
			},
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
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: &grammar.DecoratorValue{Raw: `"` + serviceDescription + `"`}},
					},
					Name:      ident("UserService"),
					Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
					Methods: []*grammar.Method{
						{
							Name: ident("getUser"),
							Input: &grammar.MethodInput{
								Arguments: []*grammar.Argument{{
									Name: ident("userId"),
									Type: plainType(grammar.String),
								}},
							},
							Output: &grammar.MethodOutput{
								Type: refGrammarType("UserProfile"),
							},
						},
					},
				},
			},
		},
	}).Model()
}

func newHashActorCredentialTestDomain(t *testing.T, credentialFieldName string) *model.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
					Vias: []*grammar.ActorVia{
						{Name: ident("client")},
					},
					Sections: []*grammar.ActorSection{
						actorAuthSection(
							[]*grammar.DataMember{{Name: ident(credentialFieldName), Type: plainType(grammar.String)}},
							[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
						),
					},
				},
			},
		},
	}).Model()
}

func newHashAllowViaTestDomain(t *testing.T, via string) *model.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
					Vias: []*grammar.ActorVia{
						{Name: ident("client")},
						{Name: ident("openapi")},
					},
				},
			},
			{
				Service: &grammar.Service{
					Name:      ident("UserService"),
					Audiences: []*grammar.ServiceAudience{serviceAllowVia("ClientActor", via)},
					Methods: []*grammar.Method{
						{Name: ident("ping")},
					},
				},
			},
			{
				Web: &grammar.Web{
					Name:      ident("UserPortalWeb"),
					Audiences: []*grammar.WebAudience{webAllowVia("ClientActor", via)},
				},
			},
		},
	}).Model()
}

func newHashTaskTestDomain(t *testing.T) *model.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{{
			Task: &grammar.Task{
				Name: ident("RebuildUserIndexTask"),
				Triggers: []*grammar.TaskTrigger{{
					Name: ident("atTime"),
					Input: &grammar.MethodInput{
						Arguments: []*grammar.Argument{{
							Name: ident("startAt"),
							Type: plainType(grammar.LocalDateTime),
						}},
					},
				}},
			},
		}},
	}).Model()
}

func newHashDataKindTestDomain(kind model.DataKind) (*model.Domain, *model.Data) {
	data := &model.Data{
		Name:     "Secret",
		SkelName: "demo.user.Secret",
		Kind:     kind,
		Members: []*model.DataMember{{
			Name: "token",
			Type: plainModelType(model.ScalarString),
		}},
	}
	spec := model.DomainSpec{Name: "demo.user"}
	switch kind {
	case model.DataKindConfig:
		data.Name = "SecretConfig"
		data.SkelName = "demo.user.SecretConfig"
		data.Lifecycle = model.ConfigLifecycleEternal
		spec.Configs = []*model.Data{data}
	case model.DataKindEvent:
		data.Name = "SecretEvent"
		data.SkelName = "demo.user.SecretEvent"
		spec.Events = []*model.Data{data}
	default:
		spec.Data = []*model.Data{data}
	}
	return model.NewDomainFromSpec(spec), data
}

func plainModelType(scalar model.Scalar) *model.Type {
	return &model.Type{Kind: model.TypeKindScalar, Scalar: scalar}
}

func analyzeHashTestDomain(t *testing.T, content *grammar.SkelContent) *analyzer.Analysis {
	t.Helper()
	analysis, diagnostics := analyzer.Analyze(content, nil)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	return analysis
}

func fillHashes(t *testing.T, domains ...*model.Domain) {
	t.Helper()
	for _, domain := range domains {
		if err := FillHashes(domain); err != nil {
			t.Fatalf("fill hashes: %v", err)
		}
	}
}

func domainContent(name string) *grammar.DomainContent {
	parts := strings.Split(name, ".")
	idents := make([]*grammar.Identifier, 0, len(parts))
	for _, part := range parts {
		idents = append(idents, ident(part))
	}
	return &grammar.DomainContent{
		Name: &grammar.QualifiedName{
			Parts: idents,
		},
	}
}

func actorAuthSection(credentialMembers []*grammar.DataMember, infoMembers []*grammar.DataMember) *grammar.ActorSection {
	return &grammar.ActorSection{
		Auth: &grammar.ActorAuth{
			Credential: &grammar.ActorCredential{Members: credentialMembers},
			Info:       &grammar.ActorInfo{Members: infoMembers},
		},
	}
}

func ident(value string) *grammar.Identifier {
	return &grammar.Identifier{Value: value}
}

func plainType(plainType grammar.PlainType) *grammar.Type {
	return &grammar.Type{Plain: &plainType}
}

func refGrammarType(name string, typeArgs ...*grammar.Type) *grammar.Type {
	return &grammar.Type{
		Reference: &grammar.ReferenceType{
			Name:          qualifiedName(name),
			TypeArguments: typeArgs,
		},
	}
}

func qualifiedName(name string) *grammar.QualifiedName {
	parts := strings.Split(name, ".")
	idents := make([]*grammar.Identifier, 0, len(parts))
	for _, part := range parts {
		idents = append(idents, ident(part))
	}
	return &grammar.QualifiedName{Parts: idents}
}

func serviceAllow(name string) *grammar.ServiceAudience {
	return &grammar.ServiceAudience{Keyword: "for", Actor: qualifiedName(name)}
}

func serviceAllowVia(name string, via string) *grammar.ServiceAudience {
	audience := serviceAllow(name)
	audience.Via = ident(via)
	return audience
}

func webAllowVia(name string, via string) *grammar.WebAudience {
	return &grammar.WebAudience{
		Keyword: "for",
		Actor:   qualifiedName(name),
		Via:     ident(via),
	}
}
