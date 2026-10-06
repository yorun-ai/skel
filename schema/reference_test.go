package schema_test

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestReferenceNamesPreserveCanonicalIdentities(t *testing.T) {
	// Programmatic schemas can contain aliases that the source analyzer rejects.
	// Canonical identities must remain stable even for those graphs.
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo",
		Imports: []*schema.Import{
			{Name: "first", Alias: "a"},
			{Name: "second", Alias: "first"},
			{Name: "other", Alias: "demo"},
			{Name: "remote", Alias: "transitive"},
		},
	})
	for input, want := range map[string]string{
		"": "", "Local": "demo.Local", "demo.Local": "demo.Local",
		"first.Value": "first.Value", "a.Value": "first.Value", "unknown.Value": "unknown.Value",
	} {
		if got := domain.ReferenceName(input); got != want {
			t.Errorf("ReferenceName(%q) = %q, want %q", input, got, want)
		}
	}
	for _, test := range []struct {
		name string
		kind *schema.Type
		want string
	}{
		{"nil", nil, ""},
		{"local", new(schema.Type{Kind: schema.TypeKindData, SkelName: "demo.Local"}), "demo.Local"},
		{"imported", new(schema.Type{Kind: schema.TypeKindEnum, SkelName: "first.State"}), "first.State"},
		{"transitive", new(schema.Type{Kind: schema.TypeKindData, SkelName: "transitive.Value"}), "transitive.Value"},
		{"data fallback", new(schema.Type{Kind: schema.TypeKindData, Data: new(schema.Data{SkelName: "demo.Local"})}), "demo.Local"},
		{"enum fallback", new(schema.Type{Kind: schema.TypeKindEnum, Enum: new(schema.Enum{SkelName: "first.State"})}), "first.State"},
		{"unqualified local", new(schema.Type{Kind: schema.TypeKindData, SkelName: "Local"}), "demo.Local"},
		{"unresolved alias", new(schema.Type{Kind: schema.TypeKindUnresolvedReference, SkelName: "Value", ExternalAlias: "a"}), "first.Value"},
		{"unresolved domain", new(schema.Type{Kind: schema.TypeKindUnresolvedReference, SkelName: "Value", ExternalAlias: "a", ExternalDomain: "first"}), "first.Value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := domain.TypeReferenceName(test.kind); got != test.want {
				t.Fatalf("TypeReferenceName = %q, want %q", got, test.want)
			}
		})
	}
}
