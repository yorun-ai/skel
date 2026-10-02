package compiler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/formatter"
	"go.yorun.ai/skelc/internal/parser"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

var benchmarkSource = []byte(`domain benchmark.demo
pub data Page<T> {
    items: list<T>
    cursor: string?
}
pub data User {
    id: uuid
    friends: list<User?>
}
resource UserResource {
    action read
}
service UserService {
    for ClientActor
    method listUsers {
        require any(UserResource:read, all(UserResource:read))
        input {
            cursor: string?
        }
        output Page<User>
    }
}
`)

func BenchmarkLexer(b *testing.B) {
	for range b.N {
		lex, err := grammar.LexerDefinition().Lex("benchmark.skel", strings.NewReader(string(benchmarkSource)))
		if err != nil {
			b.Fatal(err)
		}
		for {
			token, err := lex.Next()
			if err != nil {
				b.Fatal(err)
			}
			if token.EOF() {
				break
			}
		}
	}
}

func BenchmarkParseSource(b *testing.B) {
	for range b.N {
		if _, err := parser.ParseSource("benchmark.skel", benchmarkSource); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseSourceRecoveringManyDeclarations(b *testing.B) {
	var source strings.Builder
	source.WriteString("domain benchmark.recovery\n")
	for index := range 50 {
		fmt.Fprintf(&source, "data Value%d {\n    id string\n}\n", index)
	}
	input := []byte(source.String())
	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	b.ResetTimer()
	for range b.N {
		content, diagnostics := compiler.ParseSourceRecovering("benchmark.skel", input)
		if content == nil || len(content.Entries) != 50 || len(diagnostics) != 50 {
			b.Fatalf("unexpected recovery result: entries=%d diagnostics=%d", len(content.Entries), len(diagnostics))
		}
	}
}

func BenchmarkCheck(b *testing.B) {
	directory := writeBenchmarkDirectory(b, 40)
	b.ResetTimer()
	for range b.N {
		result, err := compiler.Check(compiler.Option{SkelIn: directory})
		if err != nil {
			b.Fatal(err)
		}
		if len(result.Diagnostics) != 0 {
			b.Fatal(result.Diagnostics)
		}
	}
}

func BenchmarkParseDirectory(b *testing.B) {
	directory := writeBenchmarkDirectory(b, 40)
	b.ResetTimer()
	for range b.N {
		if _, err := compiler.Compile(compiler.Option{SkelIn: directory}); err != nil {
			b.Fatal(err)
		}
	}
}

func writeBenchmarkDirectory(b *testing.B, count int) string {
	directory := b.TempDir()
	for index := range count {
		content := fmt.Sprintf("domain benchmark.check\npub data Value%d { id: string }\n", index)
		if err := os.WriteFile(filepath.Join(directory, fmt.Sprintf("value_%d.skel", index)), []byte(content), 0o600); err != nil {
			b.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "domain.skel"), []byte("domain benchmark.check\n"), 0o600); err != nil {
		b.Fatal(err)
	}
	return directory
}

func BenchmarkWorkspaceAnalysis(b *testing.B) {
	sources := benchmarkWorkspaceSources(40)
	b.ResetTimer()
	for range b.N {
		if diagnostics := compiler.AnalyzeWorkspace(sources); len(diagnostics) != 0 {
			b.Fatal(diagnostics)
		}
	}
}

func BenchmarkIncrementalWorkspaceAnalysis(b *testing.B) {
	sources := benchmarkWorkspaceSources(40)
	analyzer := compiler.NewWorkspaceAnalyzer()
	if diagnostics := analyzer.Analyze(sources); len(diagnostics) != 0 {
		b.Fatal(diagnostics)
	}
	b.ResetTimer()
	for range b.N {
		if diagnostics := analyzer.Analyze(sources); len(diagnostics) != 0 {
			b.Fatal(diagnostics)
		}
	}
}

func benchmarkWorkspaceSources(count int) []compiler.Source {
	sources := make([]compiler.Source, 0, count)
	for index := range count {
		name := fmt.Sprintf("benchmark.d%d", index)
		content := "domain " + name + "\npub data Value { id: string }\n"
		if index > 0 {
			previous := fmt.Sprintf("benchmark.d%d", index-1)
			content = "domain " + name + "\nimport " + previous + "\npub data Value { previous: " + previous + ".Value }\n"
		}
		sources = append(sources, compiler.Source{Path: fmt.Sprintf("/benchmark/%d.skel", index), Content: []byte(content)})
	}
	return sources
}

func FuzzParserFormatterAnalyzer(f *testing.F) {
	for _, seed := range [][]byte{
		benchmarkSource,
		[]byte("domain fuzz\ndata Node { next: Node? }\n"),
		[]byte("domain fuzz\ndata Box<T> { value: list<map<string, list<T?>>> }\n"),
		[]byte("domain fuzz\nresource File { action read }\nservice Files { method get { require any(File:read) } }\n"),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source []byte) {
		formatted, formatErr := formatter.Source(source)
		if formatErr != nil {
			return
		}
		second, formatErr := formatter.Source(formatted)
		if formatErr != nil {
			t.Fatal(formatErr)
		}
		if string(second) != string(formatted) {
			t.Fatalf("formatter is not idempotent: first=%q second=%q", formatted, second)
		}
		_, originalErr := parser.ParseSource("fuzz.skel", source)
		if originalErr == nil {
			if _, formattedErr := parser.ParseSource("fuzz.skel", formatted); formattedErr != nil {
				t.Fatalf("formatter changed valid syntax: %v", formattedErr)
			}
			originalDiagnostics := compiler.AnalyzeWorkspace([]compiler.Source{{Path: "/fuzz/input.skel", Content: source}})
			formattedDiagnostics := compiler.AnalyzeWorkspace([]compiler.Source{{Path: "/fuzz/input.skel", Content: formatted}})
			if difference := diagnosticSemanticDifference(originalDiagnostics, formattedDiagnostics); difference != "" {
				t.Fatalf("formatter changed semantic diagnostics: %s", difference)
			}
		}
	})
}

func diagnosticSemanticDifference(left, right compiler.Diagnostics) string {
	if len(left) != len(right) {
		return fmt.Sprintf("diagnostic count: original=%d formatted=%d", len(left), len(right))
	}
	for index := range left {
		if left[index].Code != right[index].Code || left[index].Severity != right[index].Severity ||
			normalizedDiagnosticMessage(left[index]) != normalizedDiagnosticMessage(right[index]) ||
			!equivalentSuggestions(left[index].Suggestion, right[index].Suggestion) ||
			!equivalentRelatedInformation(left[index].Related, right[index].Related) {
			return fmt.Sprintf("diagnostic %d: original=(%q,%q,%q) formatted=(%q,%q,%q)", index,
				left[index].Code, left[index].Severity, left[index].Message,
				right[index].Code, right[index].Severity, right[index].Message)
		}
	}
	return ""
}

func normalizedDiagnosticMessage(item compiler.Diagnostic) string {
	message := item.Message
	for _, related := range item.Related {
		message = strings.ReplaceAll(message, related.Range.Start.String(), "<related>")
	}
	return message
}

func equivalentSuggestions(left, right *compiler.DiagnosticSuggestion) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Message == right.Message && left.Replacement == right.Replacement && left.Replace == right.Replace
}

func equivalentRelatedInformation(left, right []compiler.DiagnosticRelatedInformation) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Message != right[index].Message {
			return false
		}
	}
	return true
}

func BenchmarkWorkspaceSingleFileChange(b *testing.B) {
	sources := benchmarkWorkspaceSources(40)
	analyzer := compiler.NewWorkspaceAnalyzer()
	if diagnostics := analyzer.Analyze(sources); len(diagnostics) != 0 {
		b.Fatal(diagnostics)
	}
	version := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		version++
		sources[len(sources)-1].Content = []byte(fmt.Sprintf("domain benchmark.d39\nimport benchmark.d38\npub data Value { previous: benchmark.d38.Value\nvalue%d: string }\n", version))
		if diagnostics := analyzer.Analyze(sources); len(diagnostics) != 0 {
			b.Fatal(diagnostics)
		}
	}
}
