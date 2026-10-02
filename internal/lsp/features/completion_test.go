package features

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServiceCompletesKeywordsTypesAndImportedSymbols(t *testing.T) {
	server := newFixture()
	userURI := uri.File("/workspace/user.skel")
	orderURI := uri.File("/workspace/order.skel")
	statusURI := uri.File("/workspace/status.skel")
	server.putDocument(userURI, "domain demo.user\ndata User {}\n", 1, true)
	server.putDocument(orderURI, "domain demo.order\nimport demo.user as user\ndata Order { owner: user. }\n", 1, true)
	server.putDocument(statusURI, "domain demo.order\nenum Status { ACTIVE }\n", 1, true)

	result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: orderURI},
			Position:     protocol.Position{Line: 2, Character: 25},
		},
	})
	require.NoError(t, err)
	items := result.(protocol.CompletionItemSlice)
	require.Len(t, items, 1)
	assert.Equal(t, "User", items[0].Label)

	result, err = server.service().Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: orderURI},
			Position:     protocol.Position{Line: 2, Character: 13},
		},
	})
	require.NoError(t, err)
	items = result.(protocol.CompletionItemSlice)
	assert.True(t, hasCompletion(items, "service"))
	assert.False(t, hasCompletion(items, "@sensitive"))
	assert.False(t, hasCompletion(items, "@deprecated"))
	assert.True(t, hasCompletion(items, "string"))
	assert.True(t, hasCompletion(items, "Order"))
	assert.True(t, hasCompletion(items, "Status"))
	assert.True(t, hasCompletion(items, "user"))
}

func TestServiceCompletesDecoratorPrefixForField(t *testing.T) {
	server := newFixture()
	documentURI := uri.File("/workspace/user.skel")
	source := "domain demo\nconfig FeatureFlagConfig instant {\n    @desc(\"Enabled\")\n    @e\n    enabled: bool\n}\n"
	server.putDocument(documentURI, source, 1, true)

	result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 3, Character: 6},
		},
	})
	require.NoError(t, err)
	items := result.(protocol.CompletionItemSlice)
	require.Len(t, items, 1)
	item := items[0]
	assert.Equal(t, "@example", item.Label)
	assert.Equal(t, protocol.CompletionItemKindKeyword, item.Kind)
	filterText, ok := item.FilterText.Get()
	require.True(t, ok)
	assert.Equal(t, "example", filterText)
	textEdit, ok := item.TextEdit.(*protocol.TextEdit)
	require.True(t, ok)
	assert.Equal(t, protocol.Range{
		Start: protocol.Position{Line: 3, Character: 5},
		End:   protocol.Position{Line: 3, Character: 6},
	}, textEdit.Range)
	assert.Equal(t, "example", textEdit.NewText)
}

func TestServiceCompletesDeprecatedDecoratorWithReasonSnippet(t *testing.T) {
	server := newFixture()
	server.snippetSupport = true
	documentURI := uri.File("/workspace/user.skel")
	source := "domain demo\ndata User {\n    @dep\n    legacyId: string\n}\n"
	server.putDocument(documentURI, source, 1, true)

	result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 2, Character: 8},
		},
	})
	require.NoError(t, err)
	items := result.(protocol.CompletionItemSlice)
	require.Len(t, items, 1)
	item := items[0]
	assert.Equal(t, "@deprecated", item.Label)
	assert.Equal(t, protocol.InsertTextFormatSnippet, item.InsertTextFormat)
	textEdit, ok := item.TextEdit.(*protocol.TextEdit)
	require.True(t, ok)
	assert.Equal(t, protocol.Range{
		Start: protocol.Position{Line: 2, Character: 5},
		End:   protocol.Position{Line: 2, Character: 8},
	}, textEdit.Range)
	assert.Equal(t, `deprecated("$0")`, textEdit.NewText)
}

func TestServiceFiltersDecoratorCompletionByTarget(t *testing.T) {
	tests := []struct {
		name   string
		source string
		line   uint32
		want   []string
	}{
		{
			name: "domain",
			source: "@\n" +
				"domain demo\n",
			line: 0,
			want: []string{"@desc"},
		},
		{
			name: "data declaration",
			source: "domain demo\n@\n" +
				"pub data User {}\n",
			line: 1,
			want: []string{"@deprecated", "@desc", "@sensitive"},
		},
		{
			name: "event declaration",
			source: "domain demo\n@\n" +
				"event UserCreated { payload {} }\n",
			line: 1,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name: "enum item",
			source: "domain demo\nenum Status {\n    @\n" +
				"    ACTIVE\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name:   "config field",
			source: "domain demo\nconfig AppConfig eternal {\n    @\n    value: string\n}\n",
			line:   2,
			want:   []string{"@deprecated", "@desc", "@example", "@sensitive"},
		},
		{
			name: "field",
			source: "domain demo\ndata User {\n    @\n" +
				"    password: string\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc", "@example", "@sensitive"},
		},
		{
			name: "field with description",
			source: "domain demo\ndata User {\n    @desc(\"Password\")\n    @\n" +
				"    password: string\n}\n",
			line: 3,
			want: []string{"@deprecated", "@example", "@sensitive"},
		},
		{
			name: "field with description and example",
			source: "domain demo\ndata User {\n    @desc(\"User name\")\n    @example(\"Ada\")\n    @\n" +
				"    name: string\n}\n",
			line: 4,
			want: []string{"@deprecated", "@sensitive"},
		},
		{
			name: "field with example before description",
			source: "domain demo\ndata User {\n    @example(\"Ada\")\n    @\n" +
				"    name: string\n}\n",
			line: 3,
			want: []string{"@deprecated", "@desc", "@sensitive"},
		},
		{
			name: "field with sensitive",
			source: "domain demo\ndata User {\n    @sensitive\n    @\n" +
				"    password: string\n}\n",
			line: 3,
			want: []string{"@deprecated", "@desc", "@example"},
		},
		{
			name: "event payload",
			source: "domain demo\nevent UserCreated {\n    @\n" +
				"    payload {}\n}\n",
			line: 2,
			want: []string{"@sensitive"},
		},
		{
			name: "actor credential block",
			source: "domain demo\nactor UserActor {\n    auth {\n        @\n" +
				"        credential {}\n        info {}\n    }\n}\n",
			line: 3,
			want: []string{"@sensitive"},
		},
		{
			name: "method",
			source: "domain demo\nservice UserService {\n    @\n" +
				"    method get {}\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name: "deprecated method",
			source: "domain demo\nservice UserService {\n    @deprecated(\"Use fetch\")\n    @\n" +
				"    method get {}\n}\n",
			line: 3,
			want: []string{"@desc"},
		},
		{
			name: "resource action",
			source: "domain demo\nresource User {\n    @\n" +
				"    action read\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name: "resource check",
			source: "domain demo\nresource User {\n    @\n" +
				"    check byId {}\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name: "task trigger",
			source: "domain demo\ntask RebuildTask {\n    @\n" +
				"    trigger manual {}\n}\n",
			line: 2,
			want: []string{"@deprecated", "@desc"},
		},
		{
			name:   "actor identifier field",
			source: "domain demo\nactor UserActor {\n    via client {}\n    auth {\n        credential { token: string }\n        info {\n            @\n            id: int\n        }\n    }\n}\n",
			line:   6,
			want:   []string{"@deprecated", "@desc", "@example", "@identifier", "@sensitive"},
		},
		{
			name: "input block",
			source: "domain demo\nservice UserService {\n    method get {\n        @\n" +
				"        input {}\n    }\n}\n",
			line: 3,
			want: []string{"@desc", "@sensitive"},
		},
		{
			name: "output block",
			source: "domain demo\nservice UserService {\n    method get {\n        @\n" +
				"        output string\n    }\n}\n",
			line: 3,
			want: []string{"@desc", "@example", "@sensitive"},
		},
		{
			name: "described output block",
			source: "domain demo\nservice UserService {\n    method get {\n        @desc(\"Result\")\n        @\n" +
				"        output string\n    }\n}\n",
			line: 4,
			want: []string{"@example", "@sensitive"},
		},
		{
			name: "unsupported service section",
			source: "domain demo\nservice UserService {\n    @\n" +
				"    auth\n}\n",
			line: 2,
			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newFixture()
			documentURI := uri.File("/workspace/test.skel")
			server.putDocument(documentURI, test.source, 1, true)
			result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
					Position: protocol.Position{
						Line:      test.line,
						Character: uint32(len(strings.Split(test.source, "\n")[test.line])),
					},
				},
			})
			require.NoError(t, err)
			items := result.(protocol.CompletionItemSlice)
			labels := make([]string, 0, len(items))
			for _, item := range items {
				labels = append(labels, item.Label)
			}
			assert.Equal(t, test.want, labels)
		})
	}
}

func TestServiceMarksDeprecatedCompletion(t *testing.T) {
	server := newFixture()
	documentURI := uri.File("/workspace/user.skel")
	server.putDocument(documentURI, "domain demo\n@deprecated(\"Use Profile instead\")\ndata User {}\ndata Order { user: Us }\n", 1, true)

	result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 3, Character: 21},
		},
	})
	require.NoError(t, err)
	items := result.(protocol.CompletionItemSlice)
	for _, item := range items {
		if item.Label == "User" {
			assert.Equal(t, []protocol.CompletionItemTag{protocol.CompletionItemTagDeprecated}, item.Tags)
			assert.Contains(t, string(item.Documentation.(protocol.String)), "Deprecated: Use Profile instead")
			return
		}
	}
	t.Fatal("expected User completion")
}

func TestServiceCompletesContextualLanguageValues(t *testing.T) {
	server := newFixture()
	documentURI := uri.File("/workspace/user.skel")
	source := "domain demo\nconfig DatabaseConfig et\nactor ClientActor {\n    via cl\n}\nservice UserService {\n    for ClientActor via op\n}\n"
	server.putDocument(documentURI, source, 1, true)

	tests := []struct {
		position protocol.Position
		want     []string
	}{
		{position: protocol.Position{Line: 1, Character: 24}, want: []string{"eternal", "instant"}},
		{position: protocol.Position{Line: 3, Character: 10}, want: []string{"agent", "client", "openapi"}},
		{position: protocol.Position{Line: 6, Character: 26}, want: []string{"agent", "client", "openapi"}},
	}
	for _, test := range tests {
		result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
				Position:     test.position,
			},
		})
		require.NoError(t, err)
		items := result.(protocol.CompletionItemSlice)
		labels := make([]string, 0, len(items))
		for _, item := range items {
			labels = append(labels, item.Label)
			assert.Equal(t, protocol.CompletionItemKindValue, item.Kind)
		}
		assert.Equal(t, test.want, labels)
	}
}

func hasCompletion(items protocol.CompletionItemSlice, label string) bool {
	for _, item := range items {
		if item.Label == label {
			return true
		}
	}
	return false
}

func TestServiceCompletesFullDomainQualifier(t *testing.T) {
	server := newFixture()
	userURI := uri.File("/workspace/user.skel")
	orderURI := uri.File("/workspace/order.skel")
	server.putDocument(userURI, "domain demo.user\ndata User {}\n", 1, true)
	for _, suffix := range []string{"", "Us"} {
		line := "data Order { owner: demo.user." + suffix
		server.putDocument(orderURI, "domain demo.order\nimport demo.user\n"+line, 1, true)
		result, err := server.service().Completion(t.Context(), &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: orderURI},
				Position:     protocol.Position{Line: 2, Character: uint32(len(line))},
			},
		})
		require.NoError(t, err)
		assert.True(t, hasCompletion(result.(protocol.CompletionItemSlice), "User"))
	}
}
