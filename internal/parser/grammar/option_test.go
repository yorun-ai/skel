package grammar

import (
	"testing"
)

func TestOptionsIgnoreLineComments(t *testing.T) {
	content := parseSkelForTest(t, `
// file comment
domain demo.user
service UserService { // service actor
    for ClientActor
    // method comment
    method ping {}
}
`)

	if len(content.Entries) != 1 || content.Entries[0].Service == nil {
		t.Fatalf("unexpected parsed content: %+v", content)
	}
	if content.Entries[0].Service.Name.Value != "UserService" {
		t.Fatalf("unexpected service name: %+v", content.Entries[0].Service)
	}
}

func TestOptionsIgnoreBlockComments(t *testing.T) {
	content := parseSkelForTest(t, `
/* file
comment */
domain demo.user
data User {
    /* field comment */ id: int
}
`)

	if len(content.Entries) != 1 || content.Entries[0].Data == nil {
		t.Fatalf("unexpected parsed content: %+v", content)
	}
	if content.Entries[0].Data.Name.Value != "User" {
		t.Fatalf("unexpected data name: %+v", content.Entries[0].Data)
	}
	if len(content.Entries[0].Data.Members) != 1 {
		t.Fatalf("unexpected data members: %+v", content.Entries[0].Data.Members)
	}
}
