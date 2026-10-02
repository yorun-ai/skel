package features

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/lsp/source"
)

func TestGenericBindingsHaveIndependentScopesAndRename(t *testing.T) {
	fixture := newFixture()
	documentURI := uri.File("/workspace/input.skel")
	text := "domain demo\ndata Box<TItem> { value: TItem }\ndata Other<TItem> { value: TItem }\n"
	fixture.putDocument(documentURI, text, 1, true)
	position := source.New(text).Position(strings.Index(text, "value: TItem") + len("value: "))
	location, err := fixture.service().Definition(t.Context(), &protocol.DefinitionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: position}})
	require.NoError(t, err)
	locations := location.(protocol.LocationSlice)
	require.Len(t, locations, 1)
	require.Equal(t, uint32(1), locations[0].Range.Start.Line)
	edit, err := fixture.service().Rename(t.Context(), &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: position}, NewName: "TValue"})
	require.NoError(t, err)
	require.NotNil(t, edit)
	require.Len(t, edit.Changes[documentURI], 2)
	for _, change := range edit.Changes[documentURI] {
		require.Equal(t, uint32(1), change.Range.Start.Line)
	}
	completion, err := fixture.service().Completion(t.Context(), &protocol.CompletionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: position}})
	require.NoError(t, err)
	found := false
	for _, item := range completion.(protocol.CompletionItemSlice) {
		if item.Label == "TItem" {
			found = true
		}
	}
	require.True(t, found, "generic parameter absent from scoped completion")
}

func TestRecoveredBindingAndCrossFilePrecedence(t *testing.T) {
	fixture := newFixture()
	documentURI := uri.File("/workspace/box.skel")
	fixture.putDocument(uri.File("/workspace/domain.skel"), "domain demo\n", 1, true)
	fixture.putDocument(uri.File("/workspace/value.skel"), "domain demo\ndata TItem {}\n", 1, true)
	text := "domain demo\ndata Box<TItem> { value: TItem }\ndata Broken { bad int }\n"
	fixture.putDocument(documentURI, text, 1, true)
	position := source.New(text).Position(strings.Index(text, "value: TItem") + len("value: "))
	result, err := fixture.service().Definition(t.Context(), &protocol.DefinitionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: position}})
	require.NoError(t, err)
	locations := result.(protocol.LocationSlice)
	require.Len(t, locations, 1)
	require.Equal(t, uri.File("/workspace/value.skel"), locations[0].URI, "global type must retain compiler precedence over a generic parameter")
}

func TestBindingsRemainAvailableWithoutDomainHeader(t *testing.T) {
	fixture := newFixture()
	documentURI := uri.URI("untitled:missing-domain")
	text := "data User {}\ndata Box { value: User }\n"
	fixture.putDocument(documentURI, text, 1, true)
	position := source.New(text).Position(strings.LastIndex(text, "User"))
	result, err := fixture.service().Definition(t.Context(), &protocol.DefinitionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: position}})
	require.NoError(t, err)
	locations := result.(protocol.LocationSlice)
	require.Len(t, locations, 1)
	require.Equal(t, uint32(0), locations[0].Range.Start.Line)
}
