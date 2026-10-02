package features

import (
	"fmt"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/lsp/analysis"
	"go.yorun.ai/skelc/internal/lsp/source"
)

var editingDeclarations = []string{
	`@desc("domain")
domain demo`,
	`import shared.types as shared`,
	`@desc("status")
enum Status { @desc("active") ACTIVE }`,
	`data Box<TItem> { @desc("values") values: list<map<string, TItem?>> }`,
	`config AppConfig eternal { @desc("address") address: string }`,
	`event CreatedEvent { @sensitive payload { @desc("id") id: uuid } }`,
	`actor ClientActor {
 via client {}
 auth {
 credential { @desc("token") token: string }
 info { @desc("id") id: uuid }
 }
 permission {}
 }`,
	`resource FileResource {
 @desc("read") action read {
 check owner { input { @desc("id") id: uuid } }
 }
 }`,
	`api service FileApiService {
 for ClientActor via client
 auth
 method get {
 require any(FileResource:read, all(FileResource:read:owner(id)))
 input { @desc("id") id: uuid }
 output Box<Status>
 }
 }`,
	`web FileWeb { for ClientActor via client mount /files }`,
	`task RebuildTask { trigger atTime { input { @desc("id") id: uuid } } }`,
}

func exerciseEditingSource(t *testing.T, text string) {
	t.Helper()
	documentURI := uri.File("/workspace/edit.skel")
	fixture := newFixture()
	fixture.putDocument(documentURI, text, 1, true)
	service := fixture.service()
	sources, paths := analysis.SemanticSources(service.Snapshot.DocumentsMap())
	analyzer := compiler.NewWorkspaceAnalyzer()
	if _, err := analysis.SemanticDiagnostics(t.Context(), analyzer, sources, paths); err != nil {
		t.Fatal(err)
	}
	// Exercise the fully resolved path as well as the LSP's rooted import-only path.
	for i := range sources {
		sources[i].Root = ""
	}
	if _, err := analyzer.AnalyzeContext(t.Context(), sources); err != nil {
		t.Fatal(err)
	}
	if _, err := service.DocumentSymbol(t.Context(), &protocol.DocumentSymbolParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Symbols(t.Context(), &protocol.WorkspaceSymbolParams{}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Formatting(t.Context(), &protocol.DocumentFormattingParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}}); err != nil {
		t.Fatal(err)
	}
	buffer := source.New(text)
	for _, offset := range []int{0, len(text) / 2, len(text)} {
		position := protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: documentURI}, Position: buffer.Position(offset)}
		if _, err := service.Completion(t.Context(), &protocol.CompletionParams{TextDocumentPositionParams: position}); err != nil {
			t.Fatal(err)
		}
		_, _ = service.PrepareRename(t.Context(), &protocol.PrepareRenameParams{TextDocumentPositionParams: position})
		if _, err := service.Hover(t.Context(), &protocol.HoverParams{TextDocumentPositionParams: position}); err != nil {
			t.Fatal(err)
		}
		if _, err := service.Definition(t.Context(), &protocol.DefinitionParams{TextDocumentPositionParams: position}); err != nil {
			t.Fatal(err)
		}
		if _, err := service.References(t.Context(), &protocol.ReferenceParams{TextDocumentPositionParams: position}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLanguageFeaturesHandleEditingDeclarations(t *testing.T) {
	complete := "domain demo\n" + strings.Join(editingDeclarations[2:], "\n")
	if diagnostics := compiler.AnalyzeWorkspace([]compiler.Source{{Path: "/workspace/edit.skel", Content: []byte(complete)}}); len(diagnostics) != 0 {
		t.Fatalf("editing fixture must exercise a valid resolved domain: %v", diagnostics)
	}
	declarations := append(append([]string{}, editingDeclarations...), strings.Join(editingDeclarations[2:], "\n"))
	for kind, declaration := range declarations {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			for end := 0; end <= len(declaration); end++ {
				t.Run(fmt.Sprintf("prefix-%d", end), func(t *testing.T) { exerciseEditingSource(t, "domain demo\n"+declaration[:end]) })
			}
			for offset := range len(declaration) {
				t.Run(fmt.Sprintf("delete-%d", offset), func(t *testing.T) {
					exerciseEditingSource(t, "domain demo\n"+declaration[:offset]+declaration[offset+1:])
				})
			}
		})
	}
}

func FuzzLanguageFeaturesHandleInvalidSource(f *testing.F) {
	f.Add("domain demo\n" + strings.Join(editingDeclarations[2:], "\n"))
	for _, declaration := range editingDeclarations {
		f.Add("domain demo\n" + declaration)
	}
	for _, seed := range []string{"", "domain demo\ndata User {\n @desc(\"unfinished\")\n}", "domain demo\nservice UserService { method get { require any(all("} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 4096 {
			t.Skip()
		}
		exerciseEditingSource(t, text)
	})
}
