package grammar

import (
	"strings"
	"testing"
)

func TestParseDomainContentWithTripleQuotedDescription(t *testing.T) {
	content := parseSkelForTest(t, `
@desc("""
User domain
Second line description
""")
domain demo.user
`)

	if err := content.Domain.Finalize(); err != nil {
		t.Fatalf("finalize domain skel: %v", err)
	}
	if content.Domain.Description != "User domain\nSecond line description" {
		t.Fatalf("unexpected domain description: %q", content.Domain.Description)
	}
}

func TestParseRejectsInlineTripleQuotedDescription(t *testing.T) {
	parser := buildSkelParserForTest(t)
	content, err := parser.ParseString("test.skel", strings.TrimSpace(`
@desc("""User domain
Second line description""")
domain demo.user
`))
	if err != nil {
		t.Fatalf("unexpected parse error before finalize: %v", err)
	}
	if err := content.Finalize(); err != nil {
		t.Fatalf("unexpected content finalize error: %v", err)
	}
	if err := content.Domain.Finalize(); err == nil {
		t.Fatal("expected finalize error for inline triple-quoted description")
	}
}

func TestParseIndentedTripleQuotedMethodDescription(t *testing.T) {
	content := parseSkelForTest(t, `
domain demo.user

service UserService {
    @desc("""
    Get a user by user ID
    """)
    method getUser {
        output string
    }
}
`)

	raw := serviceMethodsForTest(content.Entries[0].Service)[0].Decorators[0].Value.Raw
	got, err := UnquoteDescriptionString(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Get a user by user ID" {
		t.Fatalf("unexpected description: %q", got)
	}
}

func TestParseRejectsSingleQuotedDescription(t *testing.T) {
	parser := buildSkelParserForTest(t)
	_, err := parser.ParseString("test.skel", "domain demo.user\n@desc('User domain')\nenum Status { ACTIVE }")
	if err == nil {
		t.Fatal("expected parse error for single-quoted description")
	}
}

func TestParseSkelContentRejectsStructuralCommas(t *testing.T) {
	parser := buildSkelParserForTest(t)
	_, err := parser.ParseString("test.skel", strings.TrimSpace(`
domain demo.user

service UserService {
    method getUser {
        input {
            userId: int,
        }
    }
}
`))
	if err == nil {
		t.Fatal("expected structural comma to be rejected")
	}
}
