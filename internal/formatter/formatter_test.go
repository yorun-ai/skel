package formatter

import (
	"os"
	"path/filepath"
	"testing"

	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

func TestSourceGolden(t *testing.T) {
	input := readTestFile(t, "complete.input.skel")
	want := readTestFile(t, "complete.golden.skel")

	if _, err := parser.ParseSource("complete.input.skel", input); err != nil {
		t.Fatalf("input fixture does not parse: %v", err)
	}
	checkTestSource(t, "complete.input.skel", input)
	got := formatTestSource(t, input)
	if string(got) != string(want) {
		t.Fatalf("unexpected formatted source:\n%s\nwant:\n%s", got, want)
	}
	if _, err := parser.ParseSource("complete.golden.skel", got); err != nil {
		t.Fatalf("formatted fixture does not parse: %v", err)
	}
	checkTestSource(t, "complete.golden.skel", got)
	if second := formatTestSource(t, got); string(second) != string(got) {
		t.Fatalf("format is not idempotent:\n%s", second)
	}
}

func checkTestSource(t *testing.T, name string, source []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}
	// Like skelc --strict check, validate local semantics while allowing
	// unresolved imports, so fixtures cannot rely on migration-only syntax.
	result, err := compiler.Check(compiler.Option{SkelIn: path, Strict: true})
	if err != nil {
		t.Fatalf("check %s: %v", name, err)
	}
	if result.Diagnostics.HasErrors() {
		t.Fatalf("invalid fixture %s: %v", name, result.Diagnostics)
	}
}

func TestFormatterIsIdempotentAroundUnmatchedClosingBrace(t *testing.T) {
	first := formatTestSource(t, []byte("data}0"))
	second := formatTestSource(t, first)
	if string(first) != string(second) {
		t.Fatalf("formatter is not idempotent: first=%q second=%q", first, second)
	}
}

func TestFormatterIsIdempotentAroundMismatchedParenAndBrace(t *testing.T) {
	first := formatTestSource(t, []byte("//00\n00000(}"))
	second := formatTestSource(t, first)
	if string(first) != string(second) {
		t.Fatalf("formatter is not idempotent: first=%q second=%q", first, second)
	}
}

func TestFormatterEndsRequireSpacingAtSyntheticBlockBreak(t *testing.T) {
	first := formatTestSource(t, []byte("require{:00"))
	second := formatTestSource(t, first)
	if string(first) != string(second) {
		t.Fatalf("formatter is not idempotent: first=%q second=%q", first, second)
	}
}

func TestFormatterIsIdempotentAroundInlineTripleString(t *testing.T) {
	for _, source := range [][]byte{
		[]byte("0\"\"\"\n  \"\"\""),
		[]byte("{\"\"\"\r \r\"\"\""),
	} {
		first := formatTestSource(t, source)
		second := formatTestSource(t, first)
		if string(first) != string(second) {
			t.Errorf("formatter is not idempotent: first=%q second=%q", first, second)
		}
	}
}

func TestSourcePreservesCommentsAndStrings(t *testing.T) {
	source := []byte("domain demo.user\n\n/* comment { }\n   keep */\n@desc(\"\"\"\n  keep { content }\n    nested\n\"\"\") // inline\nservice UserService {\nmethod ping {}\n}\n")
	want := "domain demo.user\n\n/* comment { }\n   keep */\n@desc(\"\"\"\nkeep { content }\n  nested\n\"\"\") // inline\nservice UserService {\n    method ping {}\n}\n"

	got := formatTestSource(t, source)
	if string(got) != want {
		t.Fatalf("unexpected formatted source:\n%s\nwant:\n%s", got, want)
	}
	before := descriptionValue(t, source)
	after := descriptionValue(t, got)
	if before != after {
		t.Fatalf("format changed triple-string value: before=%q after=%q", before, after)
	}
}

func TestSourceRebasesBlockCommentsWithoutChangingRelativeIndentation(t *testing.T) {
	source := []byte("domain demo.user\n\ndata User {\n        /* first\n             second\n          third\n        */\nid:string\n}\n")
	want := "domain demo.user\n\ndata User {\n    /* first\n         second\n      third\n    */\n    id: string\n}\n"

	formatted := formatTestSource(t, source)
	if string(formatted) != want {
		t.Fatalf("unexpected formatted block comment:\n%s\nwant:\n%s", formatted, want)
	}
	if second := formatTestSource(t, formatted); string(second) != want {
		t.Fatalf("block comment formatting is not idempotent:\n%s", second)
	}
}

func TestSourcePreservesSemanticMetadata(t *testing.T) {
	source := []byte(`@desc("""
User domain
    with indentation
""")
domain demo.user

@desc("""
User data
    with indentation
""")
@deprecated("Use Profile instead")
@sensitive
pub data User {
@desc("User identifier")
@deprecated("Use subject instead")
@example("user-1")
@sensitive
id:string
}
`)
	formatted := formatTestSource(t, source)
	before := compileTestDomain(t, "before.skel", source)
	after := compileTestDomain(t, "after.skel", formatted)

	if before.Name() != after.Name() || before.Description() != after.Description() || before.Hash() != after.Hash() {
		t.Fatalf("format changed domain metadata: before=%q/%q/%q after=%q/%q/%q",
			before.Name(), before.Description(), before.Hash(), after.Name(), after.Description(), after.Hash())
	}
	if len(before.Data()) != 1 || len(after.Data()) != 1 {
		t.Fatalf("unexpected data declarations: before=%d after=%d", len(before.Data()), len(after.Data()))
	}
	beforeData, afterData := before.Data()[0], after.Data()[0]
	if beforeData.Description != afterData.Description || beforeData.Deprecated != afterData.Deprecated ||
		beforeData.DeprecatedReason != afterData.DeprecatedReason || beforeData.Sensitive != afterData.Sensitive {
		t.Fatalf("format changed data metadata: before=%+v after=%+v", beforeData, afterData)
	}
	if len(beforeData.Members) != 1 || len(afterData.Members) != 1 {
		t.Fatalf("unexpected data members: before=%d after=%d", len(beforeData.Members), len(afterData.Members))
	}
	beforeMember, afterMember := beforeData.Members[0], afterData.Members[0]
	if beforeMember.Description != afterMember.Description || beforeMember.Deprecated != afterMember.Deprecated ||
		beforeMember.DeprecatedReason != afterMember.DeprecatedReason || beforeMember.Example != afterMember.Example ||
		beforeMember.Sensitive != afterMember.Sensitive {
		t.Fatalf("format changed member metadata: before=%+v after=%+v", beforeMember, afterMember)
	}
}

func descriptionValue(t *testing.T, source []byte) string {
	t.Helper()
	content, err := parser.ParseSource("description.skel", source)
	if err != nil {
		t.Fatal(err)
	}
	raw := content.Entries[0].Service.Decorators[0].Value.Raw
	description, err := grammar.UnquoteDescriptionString(raw)
	if err != nil {
		t.Fatal(err)
	}
	return description
}

func TestSourcePreservesSemanticHash(t *testing.T) {
	source := []byte(`domain demo.user

pub actor ClientActor {
via client {}
}

pub data User {
id:uuid
}

pub service UserService {
for ClientActor via client
method get {
input {
id:uuid
}
output User?
}
}
`)
	formatted := formatTestSource(t, source)
	before := parseDomainHash(t, "before.skel", source)
	after := parseDomainHash(t, "after.skel", formatted)
	if before != after {
		t.Fatalf("format changed semantic hash: before=%s after=%s", before, after)
	}
}

func TestSourceNormalizesEmptyAndInvalidInput(t *testing.T) {
	if got := formatTestSource(t, []byte(" \r\n\t")); len(got) != 0 {
		t.Fatalf("expected empty output, got %q", got)
	}
	if _, err := Source([]byte("invalid !  \r\n")); err == nil {
		t.Fatal("expected invalid source error")
	}
}

func formatTestSource(t *testing.T, source []byte) []byte {
	t.Helper()
	formatted, err := Source(source)
	if err != nil {
		t.Fatal(err)
	}
	return formatted
}

func readTestFile(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func parseDomainHash(t *testing.T, name string, source []byte) string {
	return compileTestDomain(t, name, source).Hash()
}

func compileTestDomain(t *testing.T, name string, source []byte) *model.Domain {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := compiler.Compile(compiler.Option{SkelIn: path})
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return result.Domain
}

func TestActorIdentifierRoundTrip(t *testing.T) {
	source := []byte("domain demo\npub actor UserActor{via client{} auth{credential{token:string}info{\n// stable identity\n@identifier\nid:int\n}}}\n")
	before := compileTestDomain(t, "actor.skel", source)
	formatted := formatTestSource(t, source)
	after := compileTestDomain(t, "actor.skel", formatted)
	if before.Actors()[0].IdentifierField != "id" || after.Actors()[0].IdentifierField != "id" || before.Hash() != after.Hash() {
		t.Fatal("format lost actor identity metadata")
	}
	if second := formatTestSource(t, formatted); string(second) != string(formatted) {
		t.Fatal("format is not idempotent")
	}
}

func TestApiServiceRoundTrip(t *testing.T) {
	input := []byte("domain demo.order\n// client endpoint\napi   service  OrderApiService{method ping{}}\n")
	before := compileTestDomain(t, "api.skel", input)
	formatted := formatTestSource(t, input)
	after := compileTestDomain(t, "api.skel", formatted)
	if before.Hash() != after.Hash() {
		t.Fatalf("format changed API contract: %s", formatted)
	}
	parsed, err := parser.ParseSource("api.skel", formatted)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Entries[0].Service.Api || parsed.Entries[0].Service.Pub {
		t.Fatalf("lost API modifier: %s", formatted)
	}
	if string(formatTestSource(t, formatted)) != string(formatted) {
		t.Fatalf("unstable formatting: %s", formatted)
	}
}

func TestOpenServiceRoundTrip(t *testing.T) {
	input := []byte("domain demo.storage\n// reusable contract\nopen   service  StorageService{method ping{}}\n")
	before := compileTestDomain(t, "open.skel", input)
	formatted := formatTestSource(t, input)
	after := compileTestDomain(t, "open.skel", formatted)
	if !after.Services()[0].Open || before.Hash() != after.Hash() {
		t.Fatalf("lost open contract: %s", formatted)
	}
	if string(formatTestSource(t, formatted)) != string(formatted) {
		t.Fatalf("unstable format: %s", formatted)
	}
}

func TestWebMountFormattingRoundTrip(t *testing.T) {
	source := []byte("domain demo\nactor ClientActor { via client {} }\nweb PortalWeb{\nmount\t/* base */\t/portal/v1-assets/ // keep\nfor ClientActor\n}\n")
	want := "domain demo\n\nactor ClientActor {\n    via client {}\n}\n\nweb PortalWeb {\n    mount /* base */ /portal/v1-assets/ // keep\n    for ClientActor\n}\n"
	got := formatTestSource(t, source)
	if string(got) != want {
		t.Fatalf("formatted source:\n%s", got)
	}
	checkTestSource(t, "mount.skel", got)
	if second := formatTestSource(t, got); string(second) != string(got) {
		t.Fatal("format is not idempotent")
	}
}

func TestFullDomainReferencesRoundTrip(t *testing.T) {
	input := []byte("domain demo\nimport ws.sandbox\nimport other.sandbox as other\npub data Payload{value:ws.sandbox.Value alias:other.Value}\nweb ProxyWeb{for ws.sandbox.SandboxActor}\n")
	formatted := formatTestSource(t, input)
	checkTestSource(t, "domain.skel", formatted)
	if again := formatTestSource(t, formatted); string(again) != string(formatted) {
		t.Fatalf("format not idempotent: %s", again)
	}
	if _, err := parser.ParseSource("domain.skel", formatted); err != nil {
		t.Fatal(err)
	}
}
