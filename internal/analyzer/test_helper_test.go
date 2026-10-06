package analyzer

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func expectAnalyzeDiagnosticsContains(t *testing.T, expected string, content *grammar.SkelContent) {
	t.Helper()
	_, diagnostics := Analyze(content, nil)
	assertDiagnosticsContain(t, diagnostics, expected)
}

func assertDiagnosticsContain(t *testing.T, diagnostics []error, expected string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Error(), expected) {
			return
		}
	}
	t.Fatalf("expected diagnostic containing %q, got %v", expected, diagnostics)
}

func mustAnalyze(t *testing.T, content *grammar.SkelContent, importedDomains ...*Analysis) *Analysis {
	t.Helper()
	analysis, diagnostics := Analyze(content, importedDomains)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	return analysis
}

func ident(value string) *grammar.Identifier {
	return &grammar.Identifier{Value: value}
}

func plainType(pt grammar.PlainType) *grammar.Type {
	return &grammar.Type{Plain: &pt}
}

func nullableType(inner *grammar.Type) *grammar.Type {
	inner.Nullable = true
	return inner
}

func listType(value *grammar.Type) *grammar.Type {
	return &grammar.Type{
		List: &grammar.ListType{
			Value: value,
		},
	}
}

func mapType(key *grammar.Type, value *grammar.Type) *grammar.Type {
	return &grammar.Type{
		Map: &grammar.MapType{
			Key:   key,
			Value: value,
		},
	}
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

func serviceAllow(name string, via ...string) *grammar.ServiceAudience {
	audience := &grammar.ServiceAudience{Keyword: "for", Actor: qualifiedName(name)}
	if len(via) > 0 {
		audience.Via = ident(via[0])
	}
	return audience
}

func webAllow(name string, via ...string) *grammar.WebAudience {
	audience := &grammar.WebAudience{Keyword: "for", Actor: qualifiedName(name)}
	if len(via) > 0 {
		audience.Via = ident(via[0])
	}
	return audience
}

func actorWithCredentialForTest(name string) *grammar.Actor {
	return &grammar.Actor{
		Name: ident(name),
		Vias: []*grammar.ActorVia{
			{Name: ident("client")},
		},
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection(
				[]*grammar.DataMember{{Name: ident("subject"), Type: plainType(grammar.String)}},
				[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
			),
		},
	}
}

func decoratorValue(raw string) *grammar.DecoratorValue {
	return &grammar.DecoratorValue{Raw: raw}
}

func domainContent(name string) *grammar.DomainContent {
	return domainContentWithDescription(name, "")
}

func domainContentWithDescription(name string, description string) *grammar.DomainContent {
	parts := strings.Split(name, ".")
	identParts := make([]*grammar.Identifier, 0, len(parts))
	for _, part := range parts {
		identParts = append(identParts, &grammar.Identifier{Value: part})
	}
	return &grammar.DomainContent{
		Description: description,
		Name: &grammar.QualifiedName{
			Parts: identParts,
		},
	}
}
