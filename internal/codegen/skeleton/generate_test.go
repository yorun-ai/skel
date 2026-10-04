package skeleton

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/model"
)

func TestGenWritesDomainDescriptionAndOnlyNeededImports(t *testing.T) {
	_, userDir := parseDomainForTest(t, "user/domain.skel", "domain user\n", "user/types.skel", `domain user

pub data UserSummary {
    id: string
}
`, nil)
	bookerDomain, _ := parseDomainForTest(t, "booker/domain.skel", `@desc("""
Library management example
Skel definitions demonstrating catalogs, loans, tasks, and events
""")
domain booker
`, "booker/types.skel", `domain booker

import user as account

pub data Loan {
    borrower: account.UserSummary
}
`, map[string]string{"user": userDir})
	outputDir := t.TempDir()

	mustGenerateForTest(t, bookerDomain, Option{Out: outputDir, PubOnly: true})

	domainContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "domain.skel"))
	if strings.Contains(domainContent, "import user") {
		t.Fatalf("domain.skel should not include dependency imports:\n%s", domainContent)
	}
	if !strings.Contains(domainContent, `@desc("""
Library management example
Skel definitions demonstrating catalogs, loans, tasks, and events
""")
domain booker
`) {
		t.Fatalf("expected multiline domain description to stay triple quoted:\n%s", domainContent)
	}

	typesContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "types.skel"))
	if !strings.Contains(typesContent, "import user as account") {
		t.Fatalf("types.skel should include used dependency import:\n%s", typesContent)
	}
}

func TestGenKeepsDescriptionIndentation(t *testing.T) {
	domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo\n", "demo/all.skel", `domain demo

@sensitive
@desc("""
User profile
Multiline description
""")
pub data User {
    @desc("""
    User ID
    Multiline field description
    """)
    @sensitive
    id: string
}

pub data Page<TItem> {
    items: map<string, list<TItem>>
    itemsById: map<uuid, list<TItem>>
    nextToken: string?
}

@sensitive
pub config CacheConfig eternal {
    enabled: bool
}

pub event SecretEvent {
    @sensitive
    payload {
        value: string
    }
}

pub actor ClientActor {
    via client {}
    auth {
        @sensitive
        info {
            userId: string
        }
        @sensitive
        credential {
            subject: string
        }
    }
}

pub service UserService {
    for ClientActor via client

    @desc("Health check")
    method ping {
    }

    @desc("""
    Query a user
    Multiline method description
    """)
    method getUser {
        @desc("""
        Input parameters
        Multiline input description
        """)
        @sensitive
        input {
            @desc("""
            User ID
            Multiline parameter description
            """)
            @sensitive
            id: string
        }
        @desc("The returned user entity")
        @example({
id:10001,
username:"zhangsan",
displayName:"Alice",
email:"a@b.com",
})
        @sensitive
        output User
    }

    method listUsers {
        output list<User>
    }
}
`, nil)
	outputDir := t.TempDir()

	mustGenerateForTest(t, domain, Option{Out: outputDir, PubOnly: true})
	assertGeneratedSkelParses(t, outputDir)

	typesContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "types.skel"))
	assertNoExtraTopLevelBlankLines(t, typesContent)
	if !strings.Contains(typesContent, `@desc("""
User profile
Multiline description
""")
@sensitive
pub data User {
    @desc("""
    User ID
    Multiline field description
    """)
    @sensitive
    id: string
}`) {
		t.Fatalf("expected member desc indentation, got:\n%s", typesContent)
	}
	if strings.Contains(typesContent, "ClientActorCredential") || strings.Contains(typesContent, "ClientActorInfo") {
		t.Fatalf("actor implicit data should stay in actor.skel, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "pub data Page<TItem> {\n    items: map<string, list<TItem>>\n    itemsById: map<uuid, list<TItem>>\n    nextToken: string?\n}") {
		t.Fatalf("expected template-rendered generic types, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "@sensitive\npub config CacheConfig eternal {") {
		t.Fatalf("expected template-rendered config lifecycle, got:\n%s", typesContent)
	}

	eventContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "event.skel"))
	if !strings.Contains(eventContent, "pub event SecretEvent {\n    @sensitive\n    payload {") {
		t.Fatalf("expected sensitive event payload, got:\n%s", eventContent)
	}

	actorContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "actor.skel"))
	if !strings.Contains(actorContent, "    auth {\n        @sensitive\n        credential {\n            subject: string\n        }\n        @sensitive\n        info {\n            userId: string\n        }\n    }") {
		t.Fatalf("expected actor credential and info to render in stable order, got:\n%s", actorContent)
	}

	serviceContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "service.skel"))
	assertNoExtraTopLevelBlankLines(t, serviceContent)
	if !strings.Contains(serviceContent, "    for ClientActor via client\n") {
		t.Fatalf("expected service for via to render, got:\n%s", serviceContent)
	}
	if !strings.Contains(serviceContent, "    for ClientActor via client\n\n    @desc(\"Health check\")\n    method ping {}\n") {
		t.Fatalf("expected blank line after service for block, got:\n%s", serviceContent)
	}
	if !strings.Contains(serviceContent, `    @desc("""
    Query a user
    Multiline method description
    """)
    method getUser {
        @desc("""
        Input parameters
        Multiline input description
        """)
        @sensitive
        input {
            @desc("""
            User ID
            Multiline parameter description
            """)
            @sensitive
            id: string
        }
        @desc("The returned user entity")
        @example({
            id: 10001,
            username: "zhangsan",
            displayName: "Alice",
            email: "a@b.com",
        })
        @sensitive
        output User
    }`) {
		t.Fatalf("expected service desc indentation, got:\n%s", serviceContent)
	}
	if !strings.Contains(serviceContent, "    }\n\n    method listUsers {\n") {
		t.Fatalf("expected blank line between service methods, got:\n%s", serviceContent)
	}
	if strings.Contains(serviceContent, "pub service UserService {\n\n") {
		t.Fatalf("did not expect blank line before first service for, got:\n%s", serviceContent)
	}
	if !strings.Contains(serviceContent, "    method ping {}\n") {
		t.Fatalf("expected empty method to render on one line, got:\n%s", serviceContent)
	}
}

func TestGenResourceHidesPermissionCodeArgument(t *testing.T) {
	domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo\n", "demo/types.skel", `domain demo

@desc("user")
pub resource User {
    @desc("lookup check")
    check byExists {
        @desc("lookup input")
        @sensitive
        input {
            @desc("user id")
            @example(1)
            @sensitive
            userId: int
        }
    }

    @desc("read user")
    action read
    @desc("update user")
    action update {
        @desc("self check")
        check bySelf {
            input {
                @sensitive
                userId: int
            }
        }
    }
    @desc("manage user")
    action manage {
        check enabled {}
    }
}
`, nil)
	outputDir := t.TempDir()

	mustGenerateForTest(t, domain, Option{Out: outputDir, PubOnly: true})

	typesContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "types.skel"))
	if strings.Contains(typesContent, "code:") {
		t.Fatalf("resource check generated skel should hide internal code argument, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "    @desc(\"lookup check\")\n    check byExists {\n        @desc(\"lookup input\")\n        @sensitive\n        input {\n            @desc(\"user id\")\n            @example(1)\n            @sensitive\n            userId: int\n        }\n    }\n\n    @desc(\"read user\")") {
		t.Fatalf("expected blank line between resource checks and actions, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "        @desc(\"self check\")\n        check bySelf {\n            input {\n                @sensitive\n                userId: int\n            }\n        }") {
		t.Fatalf("expected action check arguments without code and with decorators, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "    }\n\n    @desc(\"manage user\")") {
		t.Fatalf("expected blank line between resource actions, got:\n%s", typesContent)
	}
	if !strings.Contains(typesContent, "        check enabled {}") {
		t.Fatalf("expected empty resource check to use an empty block, got:\n%s", typesContent)
	}
	if strings.Contains(typesContent, "        check bySelf {\n            input {\n                @sensitive\n                userId: int\n            }\n        }\n\n    }") {
		t.Fatalf("did not expect blank line before resource action closing brace, got:\n%s", typesContent)
	}
}

func TestGenDoesNotRenderBlankLineAfterServiceAudienceWithoutFollowingContent(t *testing.T) {
	domain := model.NewDomainFromSpec(model.DomainSpec{
		Name: "demo",
		Actors: []*model.Actor{
			{Name: "ClientActor", Pub: true},
		},
		Services: []*model.Service{
			{
				Name:      "EmptyService",
				Pub:       true,
				Audiences: []*model.ActorAudience{{Actor: "ClientActor"}},
			},
		},
	})
	outputDir := t.TempDir()

	mustGenerateForTest(t, domain, Option{Out: outputDir, PubOnly: true})

	serviceContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "service.skel"))
	if !strings.Contains(serviceContent, "pub service EmptyService {\n    for ClientActor\n}") {
		t.Fatalf("did not expect blank line after for without following content, got:\n%s", serviceContent)
	}
}

func parseDomainForTest(t *testing.T, domainPath string, domainContent string, inputPath string, inputContent string, imports map[string]string) (*model.Domain, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, filepath.Base(domainPath)), []byte(domainContent), 0o644); err != nil {
		t.Fatalf("write domain fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filepath.Base(inputPath)), []byte(inputContent), 0o644); err != nil {
		t.Fatalf("write input fixture: %v", err)
	}
	parsed, err := compiler.Compile(compiler.Option{SkelIn: dir, SkelImports: imports})
	if err != nil {
		t.Fatalf("parse test domain: %v", err)
	}
	return parsed.Domain, dir
}

func mustGenerateForTest(t *testing.T, domain *model.Domain, option Option) {
	t.Helper()
	if err := Generate(domain, option); err != nil {
		t.Fatalf("generate skeleton: %v", err)
	}
}

func readGeneratedFileForTest(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated file %s: %v", path, err)
	}
	return string(content)
}

func assertGeneratedSkelParses(t *testing.T, outputDir string) {
	t.Helper()
	if _, err := compiler.Compile(compiler.Option{SkelIn: outputDir}); err != nil {
		t.Fatalf("parse generated skel: %v", err)
	}
}

func assertNoExtraTopLevelBlankLines(t *testing.T, content string) {
	t.Helper()
	for _, unexpected := range []string{
		"\n\n\n@desc",
		"\n\n\npub ",
	} {
		if strings.Contains(content, unexpected) {
			t.Fatalf("unexpected extra blank line in generated skel:\n%s", content)
		}
	}
}

func TestGenPreservesDeprecatedDecorators(t *testing.T) {
	domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo\n", "demo/api.skel", `domain demo

@deprecated("Use Profile instead")
pub data User {
    @deprecated("Use id instead")
    legacyId: string
}

pub actor ClientActor {
    via client {}
}

@deprecated("Use ProfileService instead")
pub service UserService {
    for ClientActor via client

    @deprecated("Use getProfile instead")
    method getUser {
        input {
            @deprecated("Use id instead")
            legacyId: string
        }
        output User
    }
}
`, nil)
	outputDir := t.TempDir()

	mustGenerateForTest(t, domain, Option{Out: outputDir, PubOnly: true})
	assertGeneratedSkelParses(t, outputDir)

	typesContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "types.skel"))
	for _, expected := range []string{
		`@deprecated("Use Profile instead")`,
		`@deprecated("Use id instead")`,
	} {
		if !strings.Contains(typesContent, expected) {
			t.Fatalf("expected %q in generated types:\n%s", expected, typesContent)
		}
	}
	serviceContent := readGeneratedFileForTest(t, filepath.Join(outputDir, "service.skel"))
	for _, expected := range []string{
		`@deprecated("Use ProfileService instead")`,
		`@deprecated("Use getProfile instead")`,
		`@deprecated("Use id instead")`,
	} {
		if !strings.Contains(serviceContent, expected) {
			t.Fatalf("expected %q in generated service:\n%s", expected, serviceContent)
		}
	}
}

func TestPublicConfigPreservesSensitive(t *testing.T) {
	domain, _ := parseDomainForTest(t, "domain.skel", "domain demo\n", "config.skel", `domain demo
pub config TextConfig instant {
    @sensitive
    value: string?
}
`, nil)
	out := t.TempDir()
	if err := Generate(domain, Option{PubOnly: true, Out: out}); err != nil {
		t.Fatal(err)
	}
	parsed, err := compiler.Compile(compiler.Option{SkelIn: out})
	if err != nil {
		t.Fatal(err)
	}
	member := parsed.Domain.Configs()[0].Members[0]
	if !member.Sensitive {
		t.Fatalf("lost config decorators: %+v", member)
	}
}

func TestGenPreservesExtServiceModifier(t *testing.T) {
	domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo.storage\n", "demo/service.skel", `domain demo.storage
data Item { value: string }
ext service StorageService {
    method get {
        input { key: string }
        output Item
    }
}
pub service LookupService { method ping {} }
`, nil)
	output := t.TempDir()

	mustGenerateForTest(t, domain, Option{Out: output, PubOnly: true})

	parsed, err := compiler.Compile(compiler.Option{SkelIn: output, Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	services := parsed.Domain.Services()
	if len(services) != 2 || !services[1].Ext {
		t.Fatalf("public Skel lost ext modifier: %+v", services)
	}
}

func TestActorIdentifierPublicRoundTrip(t *testing.T) {
	domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo\n", "demo/actor.skel", `domain demo
pub actor UserActor {
 via client {}
 auth {
  credential { token: string }
  info { @identifier id: int }
 }
}`, nil)
	output := t.TempDir()
	mustGenerateForTest(t, domain, Option{Out: output, PubOnly: true})
	result, err := compiler.Compile(compiler.Option{SkelIn: output})
	if err != nil {
		t.Fatal(err)
	}
	if result.Domain == nil || result.Domain.Actors()[0].IdentifierField != "id" {
		t.Fatalf("lost identifier: %+v", result)
	}
}

func TestGenQualifiedImportsRoundTrip(t *testing.T) {
	_, shared := parseDomainForTest(t, "shared/domain.skel", "domain ws.sandbox\n", "shared/types.skel", `domain ws.sandbox
pub actor SandboxActor { via client {} }
pub data Value { id: string }
`, nil)
	for _, declaration := range []string{"ws.sandbox", "ws.sandbox as sandbox"} {
		t.Run(declaration, func(t *testing.T) {
			qualifier := "ws.sandbox"
			if strings.Contains(declaration, " as ") {
				qualifier = "sandbox"
			}
			domain, _ := parseDomainForTest(t, "app/domain.skel", "domain demo.proxy\n", "app/service.skel", "domain demo.proxy\nimport "+declaration+"\npub data Payload { value: "+qualifier+".Value }\npub service ProxyService { for "+qualifier+".SandboxActor method get { output Payload } }\n", map[string]string{"ws.sandbox": shared})
			out := t.TempDir()
			mustGenerateForTest(t, domain, Option{Out: out, PubOnly: true})
			_, err := compiler.Compile(compiler.Option{SkelIn: out, SkelImports: map[string]string{"ws.sandbox": shared}})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGenImportAliasesDoNotSelectUnrelatedDomains(t *testing.T) {
	imports := map[string]string{}
	for _, name := range []string{"user", "other"} {
		_, path := parseDomainForTest(t, name+"/domain.skel", "domain "+name+"\n", name+"/types.skel", "domain "+name+"\npub data Value {}\npub actor ClientActor { via client {} }\n", nil)
		imports[name] = path
	}
	for _, test := range []struct{ name, declaration, usedDomain, expectedImport, filename string }{
		{"data domain collides with alias", "pub data Payload { value: legacy.Value }", "user", "import user as legacy", "types.skel"},
		{"data alias collides with domain", "pub data Payload { value: user.Value }", "other", "import other as user", "types.skel"},
		{"actor domain collides with alias", "pub service ProxyService { for legacy.ClientActor method ping {} }", "user", "import user as legacy", "service.skel"},
		{"actor alias collides with domain", "pub service ProxyService { for user.ClientActor method ping {} }", "other", "import other as user", "service.skel"},
	} {
		t.Run(test.name, func(t *testing.T) {
			domain, _ := parseDomainForTest(t, "app/domain.skel", "domain demo.app\n", "app/content.skel", "domain demo.app\nimport user as legacy\nimport other as user\n"+test.declaration+"\n", imports)
			out := t.TempDir()
			mustGenerateForTest(t, domain, Option{Out: out, PubOnly: true})
			content := readGeneratedFileForTest(t, filepath.Join(out, test.filename))
			if strings.Count(content, "\nimport ") != 1 || !strings.Contains(content, test.expectedImport+"\n") {
				t.Fatalf("unexpected generated imports:\n%s", content)
			}
			// Consumers should only need the domain actually referenced by the contract.
			if _, err := compiler.Compile(compiler.Option{SkelIn: out, SkelImports: map[string]string{test.usedDomain: imports[test.usedDomain]}}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPublicConfigStructuredTypesRoundTrip(t *testing.T) {
	domain, _ := parseDomainForTest(t, "domain.skel", "domain demo\n", "config.skel", `domain demo
 data Entry<TValue> { value: TValue }
 data Record { content: binary children: list<Record> }
 pub config AppConfig instant {
  records: list<Record>
  entries: map<string, Entry<binary?>?>
 }
`, nil)
	out := t.TempDir()
	mustGenerateForTest(t, domain, Option{PubOnly: true, Out: out})
	parsed, err := compiler.Compile(compiler.Option{SkelIn: out})
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Domain.Data()) != 2 || len(parsed.Domain.Configs()) != 1 {
		t.Fatalf("incomplete public config dependencies: data=%d, configs=%d", len(parsed.Domain.Data()), len(parsed.Domain.Configs()))
	}
	if !parsed.Domain.Configs()[0].Members[0].Type.ContainsBinaryType() {
		t.Fatal("public config lost nested binary type")
	}
}

func TestGenPreservesAuthSyntax(t *testing.T) {
	for _, marker := range []string{"auth required", "auth optional", "auth guest", "auth", "noauth"} {
		t.Run(marker, func(t *testing.T) {
			domain, _ := parseDomainForTest(t, "demo/domain.skel", "domain demo\n", "demo/service.skel", "domain demo\npub service UserService { "+marker+" method ping { "+marker+" } }\n", nil)
			out := t.TempDir()
			mustGenerateForTest(t, domain, Option{Out: out, PubOnly: true})
			generated, err := compiler.Compile(compiler.Option{SkelIn: out})
			if err != nil {
				t.Fatal(err)
			}
			before, after := domain.Services()[0], generated.Domain.Services()[0]
			if after.Auth != before.Auth || after.Methods[0].Auth != before.Methods[0].Auth {
				t.Fatalf("auth changed during roundtrip: %q", marker)
			}
		})
	}
}
