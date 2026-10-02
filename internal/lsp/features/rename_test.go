package features

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServiceRenamesDeclarationsAndReferences(t *testing.T) {
	server := newFixture()
	userURI := uri.File("/workspace/user.skel")
	orderURI := uri.File("/workspace/order.skel")
	server.putDocument(userURI, "domain demo.user\ndata User {}\n", 1, true)
	server.putDocument(orderURI, "domain demo.order\nimport demo.user\ndata Order { owner: demo.user.User }\n", 1, true)

	edit, err := server.service().Rename(t.Context(), &protocol.RenameParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: userURI},
			Position:     protocol.Position{Line: 1, Character: 6},
		},
		NewName: "Account",
	})
	require.NoError(t, err)
	require.Len(t, edit.Changes[userURI], 1)
	require.Len(t, edit.Changes[orderURI], 1)
	assert.Equal(t, "Account", edit.Changes[userURI][0].NewText)
	assert.Equal(t, "Account", edit.Changes[orderURI][0].NewText)
}

func TestServiceRejectsInvalidRename(t *testing.T) {
	server := newFixture()
	documentURI := uri.File("/workspace/user.skel")
	server.putDocument(documentURI, "domain demo\ndata User {}\n", 1, true)

	_, err := server.service().Rename(t.Context(), &protocol.RenameParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 1, Character: 6},
		},
		NewName: "not valid",
	})
	assert.ErrorContains(t, err, "invalid Skel identifier")
}

func TestServiceDoesNotRenameUnresolvedReferences(t *testing.T) {
	server := newFixture()
	documentURI := uri.File("/workspace/user.skel")
	server.putDocument(documentURI, "domain demo\ndata User { missing: Missing }\n", 1, true)

	prepared, err := server.service().PrepareRename(t.Context(), &protocol.PrepareRenameParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 1, Character: 23},
		},
	})
	require.NoError(t, err)
	assert.Nil(t, prepared)
}

func TestRenameIsolatesIndependentDomainCopies(t *testing.T) {
	server := newFixture()
	a, b, importer := uri.File("/a/input.skel"), uri.File("/b/input.skel"), uri.File("/consumer/input.skel")
	server.putDocument(a, "domain demo\ndata User {}\n", 3, true)
	server.putDocument(b, "domain demo\ndata User {}\n", 1, true)
	server.putDocument(importer, "domain consumer\nimport demo\ndata Box { value: demo.User }\n", 1, true)
	params := &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: a}, Position: protocol.Position{Line: 1, Character: 6}}, NewName: "Account"}
	edit, err := server.service().Rename(t.Context(), params)
	require.NoError(t, err)
	require.Len(t, edit.Changes, 1)
	require.Len(t, edit.Changes[a], 1)
	// An import with two possible owning inputs cannot safely be renamed.
	params.TextDocument.URI = importer
	params.Position = protocol.Position{Line: 2, Character: 24}
	prepared, err := server.service().PrepareRename(t.Context(), &protocol.PrepareRenameParams{TextDocumentPositionParams: params.TextDocumentPositionParams})
	require.NoError(t, err)
	assert.Nil(t, prepared)
}

func TestRenameSharesDirectoryInputAndUsesDocumentVersions(t *testing.T) {
	server := newFixture()
	domain, a, b := uri.File("/a/domain.skel"), uri.File("/a/data.skel"), uri.File("/a/other.skel")
	server.putDocument(domain, "domain demo\n", 1, true)
	server.putDocument(a, "domain demo\ndata User {}\n", 7, true)
	server.putDocument(b, "domain demo\ndata Box { owner: User }\n", 0, false)
	service := server.service()
	service.DocumentChangesSupport = true
	edit, err := service.Rename(t.Context(), &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: a}, Position: protocol.Position{Line: 1, Character: 6}}, NewName: "Account"})
	require.NoError(t, err)
	require.Len(t, edit.DocumentChanges, 2)
	assert.Empty(t, edit.Changes)
	first := edit.DocumentChanges[0].(*protocol.TextDocumentEdit)
	assert.Equal(t, a, first.TextDocument.URI)
	require.NotNil(t, first.TextDocument.Version)
	assert.Equal(t, int32(7), *first.TextDocument.Version)
	second := edit.DocumentChanges[1].(*protocol.TextDocumentEdit)
	assert.Equal(t, b, second.TextDocument.URI)
	assert.Nil(t, second.TextDocument.Version)
}

func TestRenameRejectsDuplicateDeclarations(t *testing.T) {
	server := newFixture()
	a := uri.File("/a/input.skel")
	server.putDocument(a, "domain demo\ndata User {}\ndata User {}\n", 1, true)
	edit, err := server.service().Rename(t.Context(), &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: a}, Position: protocol.Position{Line: 1, Character: 6}}, NewName: "Account"})
	require.NoError(t, err)
	assert.Nil(t, edit)
}

func TestRenameRejectsParserInvalidNamesAndGenericCapture(t *testing.T) {
	for _, name := range []string{"用户", "Usér", "User١", "TItem"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture()
			u := uri.File("/workspace/input.skel")
			f.putDocument(u, "domain demo\ndata User {}\ndata Box<TItem> { value: TItem }\n", 1, true)
			edit, err := f.service().Rename(t.Context(), &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: u}, Position: protocol.Position{Line: 1, Character: 6}}, NewName: name})
			require.Error(t, err)
			assert.Nil(t, edit)
		})
	}
}

func TestRenameChecksGenericParametersInSiblingFiles(t *testing.T) {
	f := newFixture()
	u := uri.File("/workspace/user.skel")
	f.putDocument(uri.File("/workspace/domain.skel"), "domain demo\n", 1, true)
	f.putDocument(u, "domain demo\ndata User {}\n", 1, true)
	f.putDocument(uri.File("/workspace/box.skel"), "domain demo\ndata Box<TItem> { value: TItem }\n", 1, true)
	edit, err := f.service().Rename(t.Context(), &protocol.RenameParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: u}, Position: protocol.Position{Line: 1, Character: 6}}, NewName: "TItem"})
	require.Error(t, err)
	assert.Nil(t, edit)
}
