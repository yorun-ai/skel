package schema

import (
	"fmt"
	"testing"
)

func ExampleNewDomainFromSpec() {
	domain := NewDomainFromSpec(DomainSpec{
		Name: "demo.user",
		Data: []*Data{
			{Name: "User", SkelName: "demo.user.User"},
		},
	})

	fmt.Println(domain.Name())
	fmt.Println(domain.Data()[0].SkelName)

	// Output:
	// demo.user
	// demo.user.User
}

func TestNewDomainFromSpecStoresSemanticData(t *testing.T) {
	data := &Data{Name: "User"}
	domain := NewDomainFromSpec(DomainSpec{
		Name:        "demo.user",
		Description: "User domain",
		Hash:        "12345678",
		Data:        []*Data{data},
	})

	if domain.Name() != "demo.user" || domain.Description() != "User domain" || domain.Hash() != "12345678" {
		t.Fatalf("unexpected domain metadata: name=%q description=%q hash=%q", domain.Name(), domain.Description(), domain.Hash())
	}
	if len(domain.Data()) != 1 || domain.Data()[0] != data {
		t.Fatalf("unexpected domain data: %+v", domain.Data())
	}
}
