package features

import (
	"go.lsp.dev/uri"
	"go.yorun.ai/skel/internal/lsp/workspace"
)

type testFixture struct {
	workspace      *workspace.Store
	snippetSupport bool
}

func newFixture() *testFixture {
	return &testFixture{workspace: workspace.New()}
}

func (f *testFixture) putDocument(documentURI uri.URI, content string, version int32, open bool) {
	f.workspace.Put(documentURI, content, version, open)
}

func (f *testFixture) service() *Service {
	return &Service{Snapshot: f.workspace.Snapshot(), SnippetSupport: f.snippetSupport}
}
