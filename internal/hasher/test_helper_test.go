package hasher

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/analyzer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func newHashTestDomain(t *testing.T, serviceDescription string) *schema.Domain {
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
				},
			},
			{
				Service: &grammar.Service{
					Auth: &grammar.AuthMarker{Value: "required"},
					Api:  true,
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: &grammar.DecoratorValue{Raw: `"` + serviceDescription + `"`}},
					},
					Name:      ident("UserApiService"),
					Audiences: []*grammar.ActorAudience{actorAudience("ClientActor")},
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
	}).Schema()
}

func newHashActorCredentialTestDomain(t *testing.T, credentialFieldName string) *schema.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
					Sections: []*grammar.ActorSection{
						actorAuthSection(
							[]*grammar.DataMember{{Name: ident(credentialFieldName), Type: plainType(grammar.String)}},
							[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
						),
					},
				},
			},
		},
	}).Schema()
}

// TODO: Remove with the legacy via compatibility tests when its syntax and warnings are removed.
func newHashAllowViaTestDomain(t *testing.T, via string) *schema.Domain {
	return analyzeHashTestDomain(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Actor: &grammar.Actor{
					Name: ident("ClientActor"),
					Sections: []*grammar.ActorSection{
						{Via: new(grammar.ActorVia{Name: ident("client")})},
						{Via: new(grammar.ActorVia{Name: ident("openapi")})},
					},
				},
			},
			{
				Service: &grammar.Service{
					Auth:      &grammar.AuthMarker{Value: "required"},
					Api:       true,
					Name:      ident("UserApiService"),
					Audiences: []*grammar.ActorAudience{actorAudience("ClientActor", via)},
					Methods: []*grammar.Method{
						{Name: ident("ping")},
					},
				},
			},
			{
				Web: &grammar.Web{
					Sections:  []*grammar.WebSection{{Auth: &grammar.AuthMarker{Value: "required"}}},
					Name:      ident("UserPortalWeb"),
					Audiences: []*grammar.ActorAudience{actorAudience("ClientActor", via)},
				},
			},
		},
	}).Schema()
}

func newHashTaskTestDomain(t *testing.T) *schema.Domain {
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
	}).Schema()
}

func newHashDataKindTestDomain(kind schema.DataKind) (*schema.Domain, *schema.Data) {
	data := &schema.Data{
		Name:     "Secret",
		SkelName: "demo.user.Secret",
		Kind:     kind,
		Members: []*schema.DataMember{{
			Name: "token",
			Type: plainSchemaType(schema.ScalarString),
		}},
	}
	spec := schema.DomainSpec{Name: "demo.user"}
	switch kind {
	case schema.DataKindConfig:
		data.Name = "SecretConfig"
		data.SkelName = "demo.user.SecretConfig"
		data.Lifecycle = schema.ConfigLifecycleEternal
		spec.Configs = []*schema.Data{data}
	case schema.DataKindEvent:
		data.Name = "SecretEvent"
		data.SkelName = "demo.user.SecretEvent"
		spec.Events = []*schema.Data{data}
	default:
		spec.Data = []*schema.Data{data}
	}
	return schema.NewDomainFromSpec(spec), data
}

func plainSchemaType(scalar schema.Scalar) *schema.Type {
	return &schema.Type{Kind: schema.TypeKindScalar, Scalar: scalar}
}

func analyzeHashTestDomain(t *testing.T, content *grammar.SkelContent) *analyzer.Analysis {
	t.Helper()
	analysis, diagnostics := analyzer.Analyze(content, nil)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	return analysis
}

func fillHashes(t *testing.T, domains ...*schema.Domain) {
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

// TODO: Drop the via argument when legacy via syntax and warnings are removed.
func actorAudience(name string, via ...string) *grammar.ActorAudience {
	entry := new(grammar.ActorAudienceEntry{Actor: qualifiedName(name)})
	if len(via) > 0 {
		entry.Via = ident(via[0])
	}
	return new(grammar.ActorAudience{Keyword: "for", Actors: []*grammar.ActorAudienceEntry{entry}})
}
