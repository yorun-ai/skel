package codegen

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestBuildValidatesPublicActorCredentialClosure(t *testing.T) {
	privateData := &schema.Data{Name: "Secret", Kind: schema.DataKindData}
	credential := &schema.Data{Name: "Credential", Members: []*schema.DataMember{{
		Name: "secret", Type: &schema.Type{Kind: schema.TypeKindData, Data: privateData},
	}}}
	domain := schema.NewDomainFromSpec(schema.DomainSpec{
		Name: "demo.user", Data: []*schema.Data{privateData},
		Actors: []*schema.Actor{{Pub: true, Name: "UserActor", Auth: new(schema.ActorAuth{Credential: credential, Info: &schema.Data{Name: "Info"}})}},
	})
	view, err := BuildPublicView(domain)
	if err != nil || len(view.Data) != 1 || view.Data[0] != privateData {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
