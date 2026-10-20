package schema

import (
	"fmt"
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
