package compiler

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/analyzer"
	"go.yorun.ai/skel/schema"
)

func TestParseSourceRecoveringCollectsIndependentSyntaxErrors(t *testing.T) {
	source := []byte(`domain demo
data User {
    first string
    second:
    third: string
}
data Order {
    id string
}
`)
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", source)
	if content == nil {
		t.Fatal("expected recovered syntax tree")
	}
	if len(diagnostics) != 3 {
		t.Fatalf("expected three syntax diagnostics, got %d: %v", len(diagnostics), diagnostics)
	}
	if len(content.Entries) != 2 {
		t.Fatalf("expected both top-level declarations to survive recovery, got %d", len(content.Entries))
	}
	if got := content.Entries[0].Data.Members; len(got) != 1 || got[0].Name.Value != "third" {
		t.Fatalf("unexpected recovered members: %+v", got)
	}
}

func TestWorkspaceReportsRecoveredSyntaxAndSemanticDiagnostics(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/domain.skel",
		Content: []byte(`domain demo
data User {
    id string
}
data User {}
`),
	}})
	if len(diagnostics) != 2 {
		t.Fatalf("expected syntax and semantic diagnostics, got %d: %v", len(diagnostics), diagnostics)
	}
	if !strings.HasPrefix(diagnostics[0].Code, "syntax.") || diagnostics[1].Code != DiagnosticCodeSemanticDuplicate {
		t.Fatalf("unexpected diagnostic codes: %s, %s", diagnostics[0].Code, diagnostics[1].Code)
	}
	if len(diagnostics[1].Related) != 1 {
		t.Fatalf("expected duplicate related location, got %+v", diagnostics[1].Related)
	}
}

func TestWorkspaceAnalysisHandlesIncompleteNestedRequire(t *testing.T) {
	diagnostics := AnalyzeWorkspace([]Source{{
		Path: "/workspace/service.skel",
		Content: []byte(`domain demo
service UserService {
    require any(all(UserResource:read
`),
	}})

	if !slices.ContainsFunc(diagnostics, func(diagnostic Diagnostic) bool {
		return strings.HasPrefix(diagnostic.Code, "syntax.")
	}) {
		t.Fatalf("expected syntax diagnostic, got %v", diagnostics)
	}
}

func TestSourceRangeUsesRuneColumns(t *testing.T) {
	tests := []struct {
		source    string
		start     schema.Position
		endColumn int
	}{
		{source: `@desc("𐐀") data user {}`, start: schema.Position{File: "/workspace/data.skel", Line: 1, Column: 17}, endColumn: 21},
		{source: `😀`, start: schema.Position{File: "/workspace/data.skel", Line: 1, Column: 1}, endColumn: 2},
	}
	for _, test := range tests {
		range_ := sourceRangeAt(test.start, []byte(test.source))
		if range_.Start != test.start || range_.End.Column != test.endColumn {
			t.Errorf("unexpected Unicode-aware source range for %q: %+v", test.source, range_)
		}
	}
}

func TestParseSourceRecoveringSynchronizesAtMissingBlockEnd(t *testing.T) {
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", []byte(`domain demo
data User {
    id: string
data Order {
    id: string
}
`))
	if content == nil || len(content.Entries) != 2 {
		t.Fatalf("expected recovery at next top-level declaration: content=%+v diagnostics=%v", content, diagnostics)
	}
	if len(diagnostics) == 0 || diagnostics[0].Suggestion == nil || diagnostics[0].Suggestion.Replacement != "}\n" {
		t.Fatalf("expected missing block-end suggestion, got %v", diagnostics)
	}
}

func TestParseSourceRecoveringKeepsDecoratorAfterMissingBlockEnd(t *testing.T) {
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", []byte(`domain demo
data User {
    id: string
@desc("order")
data Order {}
`))
	if content == nil || len(content.Entries) != 2 {
		t.Fatalf("expected both declarations, got content=%+v diagnostics=%v", content, diagnostics)
	}
	order := content.Entries[1].Data
	if order == nil || len(order.Decorators) != 1 || order.Decorators[0].Name.Value != "desc" {
		t.Fatalf("expected decorator to remain attached to Order: %+v", order)
	}
	if len(diagnostics) != 1 || diagnostics[0].Suggestion == nil || diagnostics[0].Suggestion.Replacement != "}\n" {
		t.Fatalf("expected one missing block-end diagnostic, got %v", diagnostics)
	}
}

func TestParseSourceRecoveringSynchronizesAfterBrokenDecorator(t *testing.T) {
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", []byte(`domain demo
@desc(
data User {}
`))
	if content == nil || len(content.Entries) != 1 || content.Entries[0].Data == nil {
		t.Fatalf("expected declaration after decorator to survive: content=%+v diagnostics=%v", content, diagnostics)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != DiagnosticCodeSyntaxEOF {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
}

func TestParseSourceRecoveringKeepsTopLevelKeywordsInsideDeclarations(t *testing.T) {
	source := []byte(`domain demo
data Example {
    data: string
    @desc("""first line
data NotADeclaration {}
last line""")
    description: string
}
data Actual {}
`)
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", source)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
	if content == nil || len(content.Entries) != 2 {
		t.Fatalf("expected two declarations, got %+v", content)
	}
	if got := content.Entries[0].Data.Members; len(got) != 2 || got[0].Name.Value != "data" {
		t.Fatalf("unexpected data members: %+v", got)
	}
}

func TestParseSourceRecoveringPreservesFragmentPositionsAndDecorators(t *testing.T) {
	source := []byte(`domain demo

@desc("order")
data Order {
    id string
}
`)
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", source)
	if content == nil || len(content.Entries) != 1 {
		t.Fatalf("expected recovered declaration, got %+v", content)
	}
	entry := content.Entries[0]
	if entry.Pos.Line != 3 || entry.Data.Pos.Line != 4 || len(entry.Data.Decorators) != 1 {
		t.Fatalf("fragment positions or decorators changed: %+v", entry)
	}
	if len(diagnostics) != 1 || diagnostics[0].Position.Line != 5 {
		t.Fatalf("expected syntax error on original line 5, got %v", diagnostics)
	}
}

func TestParseSourceRecoveringPreservesDeclarationOrderingRules(t *testing.T) {
	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", []byte(`domain demo
data User {}
import shared.types
domain other
data Order {}
`))
	if content == nil || len(content.Entries) != 2 || len(content.Imports) != 0 {
		t.Fatalf("unexpected recovered content: %+v", content)
	}
	if len(diagnostics) != 2 {
		t.Fatalf("expected import and domain ordering diagnostics, got %v", diagnostics)
	}
	if diagnostics[0].Position.Line != 3 || diagnostics[1].Position.Line != 4 {
		t.Fatalf("unexpected diagnostic positions: %v", diagnostics)
	}
}

func TestParseSourceRecoveringRetainsDeclarationsAfterDiagnosticLimit(t *testing.T) {
	var source strings.Builder
	source.WriteString("domain demo\n")
	for index := range analyzer.MaxDiagnosticsPerDomain {
		fmt.Fprintf(&source, "data Broken%d { id string }\n", index)
	}
	source.WriteString("data Valid { id: string }\n")

	content, diagnostics := ParseSourceRecovering("/workspace/domain.skel", []byte(source.String()))
	if len(diagnostics) != analyzer.MaxDiagnosticsPerDomain {
		t.Fatalf("expected diagnostics to stop at %d, got %d", analyzer.MaxDiagnosticsPerDomain, len(diagnostics))
	}
	if content == nil || len(content.Entries) == 0 {
		t.Fatalf("expected recovered declarations, got %+v", content)
	}
	last := content.Entries[len(content.Entries)-1].Data
	if last == nil || last.Name.Value != "Valid" {
		t.Fatalf("expected declaration after diagnostic limit to survive, got %+v", last)
	}
}

func TestWorkspaceAnalysisDiscardsUnrecoverableFragments(t *testing.T) {
	for _, declaration := range []string{
		"config AppConfig eternal {\n    @desc(\"unfinished field\")",
		"data User {\n    @desc(\"unfinished field\")\n}",
		"service UserService { method",
		"data",
	} {
		t.Run(declaration, func(t *testing.T) {
			source := []byte("domain demo\n" + declaration + "\ndata Duplicate {}\ndata Duplicate {}\n")
			diagnostics := AnalyzeWorkspace([]Source{{Path: "/workspace/incomplete.skel", Content: source}})
			if !slices.ContainsFunc(diagnostics, func(diagnostic Diagnostic) bool {
				return strings.HasPrefix(diagnostic.Code, "syntax.")
			}) {
				t.Fatalf("expected syntax diagnostic, got %v", diagnostics)
			}
			if !slices.ContainsFunc(diagnostics, func(diagnostic Diagnostic) bool {
				return diagnostic.Code == DiagnosticCodeSemanticDuplicate
			}) {
				t.Fatalf("expected later declarations to receive semantic analysis, got %v", diagnostics)
			}
		})
	}
}

func TestWorkspaceAnalysisHandlesIncompleteFragmentAfterDiagnosticLimit(t *testing.T) {
	var source strings.Builder
	source.WriteString("domain demo\n")
	for index := range analyzer.MaxDiagnosticsPerDomain {
		fmt.Fprintf(&source, "data Broken%d { id string }\n", index)
	}
	source.WriteString("config AppConfig eternal {\n    @desc(\"unfinished field\")\n}\ndata Valid {}\n")
	diagnostics := AnalyzeWorkspace([]Source{{Path: "/workspace/limited.skel", Content: []byte(source.String())}})
	if len(diagnostics) != analyzer.MaxDiagnosticsPerDomain {
		t.Fatalf("expected %d diagnostics, got %v", analyzer.MaxDiagnosticsPerDomain, diagnostics)
	}
}

func TestWorkspaceAnalysisHandlesEveryConfigEditingPrefix(t *testing.T) {
	declaration := `config AppConfig eternal {
    @desc("server address")
    address: string
}`
	for end := 0; end <= len(declaration); end++ {
		t.Run(fmt.Sprint(end), func(t *testing.T) {
			source := []byte("domain demo\n" + declaration[:end])
			diagnostics := AnalyzeWorkspace([]Source{{Path: "/workspace/config.skel", Content: source}})
			if end == len(declaration) && len(diagnostics) != 0 {
				t.Fatalf("expected valid complete config, got %v", diagnostics)
			}
		})
	}
}

func TestRecoveryContextCancelsWithoutPartialSyntax(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	content, diagnostics, err := ParseSourceRecoveringContext(ctx, "input.skel", []byte("domain demo\ndata User {"))
	if !errors.Is(err, context.Canceled) || content != nil || diagnostics != nil {
		t.Fatalf("unexpected cancelled recovery: %v %v %v", content, diagnostics, err)
	}
}
