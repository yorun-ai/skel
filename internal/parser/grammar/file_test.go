package grammar

import (
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestTopLevelDeclarationKeywordsMatchSkelEntryGrammar(t *testing.T) {
	quotedKeyword := regexp.MustCompile(`"([a-z]+)"`)
	entryType := reflect.TypeFor[SkelEntry]()
	keywords := make([]string, 0, entryType.NumField())
	for index := range entryType.NumField() {
		field := entryType.Field(index)
		if field.Name == "Pos" || field.Name == "Decorators" || field.Name == "Pub" || field.Name == "Api" || field.Name == "Ext" {
			continue
		}
		match := quotedKeyword.FindStringSubmatch(field.Tag.Get("parser"))
		if len(match) != 2 {
			t.Fatalf("declaration field %s has no keyword in parser tag %q", field.Name, field.Tag.Get("parser"))
		}
		keywords = append(keywords, match[1])
	}
	if !slices.Equal(keywords, TopLevelDeclarationKeywords()) {
		t.Fatalf("declaration keywords = %q, want %q", TopLevelDeclarationKeywords(), keywords)
	}
}

func TestParseDomainContent(t *testing.T) {
	content := parseSkelForTest(t, `
@desc("User domain")
domain demo.user
`)

	if err := content.Domain.Finalize(); err != nil {
		t.Fatalf("finalize domain skel: %v", err)
	}
	if content.Domain == nil || content.Domain.Name == nil {
		t.Fatalf("unexpected domain content: %+v", content.Domain)
	}
	if content.Domain.Name.String() != "demo.user" {
		t.Fatalf("unexpected domain name: %s", content.Domain.Name.String())
	}
	if content.Domain.Description != "User domain" {
		t.Fatalf("unexpected domain description: %q", content.Domain.Description)
	}
}

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
