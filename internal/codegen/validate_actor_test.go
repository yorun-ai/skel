package codegen

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestPrepareActorInfoRequiresMembersButNotIdentifier(t *testing.T) {
	for _, imported := range []bool{false, true} {
		for _, tc := range []struct {
			name    string
			members []*schema.DataMember
			valid   bool
		}{
			{name: "nil members"},
			{name: "empty members", members: []*schema.DataMember{}},
			{name: "without identifier", members: []*schema.DataMember{{Name: "role", Type: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString})}}, valid: true},
			{name: "nullable field", members: []*schema.DataMember{{Name: "role", Type: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString, Nullable: true})}}, valid: true},
		} {
			name := tc.name
			if imported {
				name = "imported/" + name
			}
			t.Run(name, func(t *testing.T) {
				info := new(schema.Data{Name: "UserActorInfo", Kind: schema.DataKindData, Members: tc.members})
				method := new(schema.Method{Name: "auth", ResultType: new(schema.Type{Kind: schema.TypeKindData, Data: info})})
				actor := new(schema.Actor{
					Name: "UserActor",
					Auth: new(schema.ActorAuth{
						Credential: new(schema.Data{
							Name:    "UserActorCredential",
							Kind:    schema.DataKindData,
							Members: []*schema.DataMember{{Name: "token", Type: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString})}},
						}),
						Info:    info,
						Method:  method,
						Service: new(schema.Service{Name: "UserActorAuthService", Methods: []*schema.Method{method}}),
					}),
				})
				domain := schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo", Actors: []*schema.Actor{actor}})
				if imported {
					domain = schema.NewDomainFromSpec(schema.DomainSpec{Name: "app", Imports: []*schema.Import{{Name: "demo", Domain: domain}}})
				}
				_, err := Prepare(domain, Selection{})
				if tc.valid {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), "auth info must have at least one member") {
					t.Fatalf("expected empty info error, got %v", err)
				}
			})
		}
	}
}
