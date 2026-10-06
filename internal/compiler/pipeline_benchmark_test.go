package compiler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/internal/parser/grammar"
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

func BenchmarkCompileDirectory(b *testing.B) {
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
		sources[len(sources)-1].Content = fmt.Appendf(nil, "domain benchmark.d39\nimport benchmark.d38\npub data Value { previous: benchmark.d38.Value\nvalue%d: string }\n", version)
		if diagnostics := analyzer.Analyze(sources); len(diagnostics) != 0 {
			b.Fatal(diagnostics)
		}
	}
}
