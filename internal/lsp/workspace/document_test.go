package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	lspsource "go.yorun.ai/skel/internal/lsp/source"
	"go.yorun.ai/skel/internal/schema"
)

func TestIndexDocumentDefinitionsAndReferences(t *testing.T) {
	source := `domain demo.order
import demo.user as user

// user.Ignored and Ignored must not be indexed.
@desc("user.Ignored")
data Order {
    owner: user.User
    parent: Order?
}
`
	document := BuildDocument(uri.File("/workspace/order.skel"), "/workspace/order.skel", source, 1)
	require.Len(t, document.Definitions, 1)
	assert.Equal(t, "demo.order.Order", document.Definitions[0].Key)

	keys := make([]string, 0, len(document.Occurrences))
	for _, occurrence := range document.Occurrences {
		keys = append(keys, occurrence.Key)
	}
	assert.Equal(t, []string{"demo.order.Order", "demo.user.User", "demo.order.Order"}, keys)
}

func TestIndexDocumentCoversEverySchemaDeclarationType(t *testing.T) {
	document := BuildDocument(uri.File("/workspace/all.skel"), "/workspace/all.skel", `domain demo
actor Caller { via client {} }
config Runtime eternal {}
data Record {}
enum State { READY }
event Changed { payload {} }
resource Document {}
service Documents {}
task Cleanup {}
web Portal {}
`, 1)
	require.Empty(t, document.ParseDiagnostics)
	details := make([]string, 0, len(document.Definitions))
	for _, definition := range document.Definitions {
		details = append(details, definition.Detail)
	}
	expected := make([]string, 0, len(schema.DeclarationTypes()))
	for _, kind := range schema.DeclarationTypes() {
		expected = append(expected, string(kind))
	}
	assert.ElementsMatch(t, expected, details)
}

func TestIndexDocumentUsesUTF16Positions(t *testing.T) {
	source := "domain demo\n@desc(\"𐐀\") data User {}\n"
	document := BuildDocument(uri.File("/workspace/user.skel"), "/workspace/user.skel", source, 1)
	require.Len(t, document.Definitions, 1)
	assert.Equal(t, protocol.Position{Line: 1, Character: 17}, document.Definitions[0].Range.Start)
}

func TestIndexDocumentKeepsSyntaxError(t *testing.T) {
	document := BuildDocument(uri.File("/workspace/invalid.skel"), "/workspace/invalid.skel", "domain demo\ndata User {", 1)
	require.Len(t, document.ParseDiagnostics, 1)
	assert.Equal(t, "syntax.unexpected-eof", document.ParseDiagnostics[0].Code)
	require.Len(t, document.Definitions, 1)
	assert.Equal(t, "demo.User", document.Definitions[0].Key)
	require.Len(t, document.Occurrences, 1)
	assert.Equal(t, "demo.User", document.Occurrences[0].Key)
}

func TestIndexDocumentBuildsNestedSymbols(t *testing.T) {
	source := `domain demo
service UserService {
    method getUser {
        input {
            userId: int
        }
        output string
    }
}
`
	document := BuildDocument(uri.File("/workspace/service.skel"), "/workspace/service.skel", source, 1)
	require.Len(t, document.Symbols, 1)
	service := document.Symbols[0]
	assert.Equal(t, "UserService", service.Name)
	require.Len(t, service.Children, 1)
	method := service.Children[0]
	assert.Equal(t, "getUser", method.Name)
	require.Len(t, method.Children, 1)
	assert.Equal(t, "userId", method.Children[0].Name)
	assert.LessOrEqual(t, lspsource.ComparePosition(method.Children[0].Range.End, method.Range.End), 0)
}

func TestIndexExtService(t *testing.T) {
	document := BuildDocument(uri.File("/workspace/ext.skel"), "/workspace/ext.skel", "domain demo\next service StorageService { method ping {} }\n", 1)
	require.Len(t, document.Symbols, 1)
	assert.Equal(t, "ext service", document.Symbols[0].Detail)
	assert.Equal(t, "StorageService", document.Symbols[0].Name)
}

func TestIndexExtEvent(t *testing.T) {
	document := BuildDocument(uri.File("/workspace/ext.skel"), "/workspace/ext.skel", "domain demo\next event AuditRecordedEvent { payload { message: string } }\n", 1)
	require.Len(t, document.Symbols, 1)
	assert.Equal(t, "ext event", document.Symbols[0].Detail)
	assert.Equal(t, "AuditRecordedEvent", document.Symbols[0].Name)
	require.Len(t, document.Symbols[0].Children, 1)
}

func TestOccurrencesOnlyIncludeGrammarReferences(t *testing.T) {
	text := `domain demo
import demo.other as other
data User {}
data Box {
    User: string
    value: list<map<string, User?>>
    external: other.User
}
service UserService {
    method User {
        input { User: User }
        output User
    }
}
`
	document := BuildDocument(uri.File("/audit/input.skel"), "/audit/input.skel", text, 1)
	require.Empty(t, document.ParseDiagnostics)
	keys := []string{}
	for _, occurrence := range document.Occurrences {
		keys = append(keys, occurrence.Key)
	}
	assert.Equal(t, []string{"demo.User", "demo.Box", "demo.User", "demo.other.User", "demo.UserService", "demo.User", "demo.User"}, keys)
}

func TestIncompleteOccurrencesRemainConservative(t *testing.T) {
	document := BuildDocument(uri.File("/audit/input.skel"), "/audit/input.skel", "domain demo\ndata User {}\ndata Box {\n    User: string\n    owner: User\n", 1)
	require.NotEmpty(t, document.ParseDiagnostics)
	occurrences := 0
	for _, value := range document.Occurrences {
		if value.Key == "demo.User" {
			occurrences++
		}
	}
	assert.Equal(t, 2, occurrences) // declaration and recovered type, never the field
}

func TestRecoveryDoesNotTreatDecoratorWordsAsDeclarations(t *testing.T) {
	document := BuildDocument(uri.File("/audit/input.skel"), "/audit/input.skel", "domain demo\n@desc(data Phantom {})\ndata User {}\ndata Broken {", 1)
	require.NotEmpty(t, document.ParseDiagnostics)
	for _, occurrence := range document.Occurrences {
		assert.NotEqual(t, "demo.Phantom", occurrence.Key)
	}
	for _, definition := range document.Definitions {
		if definition.Name == "Phantom" {
			assert.False(t, definition.Confirmed)
		}
	}
}
