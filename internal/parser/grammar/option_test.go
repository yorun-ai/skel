package grammar

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
)

func TestOptionsSupportDecorator(t *testing.T) {
	parser := buildSkelParserForTest(t)
	src := strings.TrimSpace(`
domain demo.user

service AgentService {
    for AgentActor

    method call {
        input {
            prompt: string
        }
    }
}
`)

	content, err := parser.ParseString("agent.skel", src, participle.AllowTrailing(true))
	if err != nil {
		t.Fatalf("parse with options: %v", err)
	}
	if err := content.Finalize(); err != nil {
		t.Fatalf("finalize content: %v", err)
	}
	if len(content.Entries) != 1 || content.Entries[0].Service == nil {
		t.Fatalf("unexpected parsed content: %+v", content)
	}
	audiences := serviceAudiencesForTest(content.Entries[0].Service)
	methods := serviceMethodsForTest(content.Entries[0].Service)
	input := methodInputForTest(methods[0])
	if audiences[0].Actors[0].Actor.String() != "AgentActor" {
		t.Fatalf("unexpected actor parse result: %+v", audiences)
	}
	if input == nil {
		t.Fatalf("unexpected method input: %+v", methods[0])
	}
	if len(input.Arguments) != 1 {
		t.Fatalf("unexpected method args: %+v", input.Arguments)
	}
}

func TestOptionsLexerRecognizesAtSign(t *testing.T) {
	parser := buildSkelParserForTest(t)
	content, err := parser.ParseString("decorator.skel", "domain demo.user\nservice UserService { for ClientActor\nmethod ping {} }")
	if err != nil {
		t.Fatalf("parse decorator skel: %v", err)
	}
	if err := content.Finalize(); err != nil {
		t.Fatalf("finalize content: %v", err)
	}
	audiences := serviceAudiencesForTest(content.Entries[0].Service)
	if audiences[0].Actors[0].Actor.String() != "ClientActor" {
		t.Fatalf("unexpected actor tokenization: %+v", audiences)
	}
}

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
