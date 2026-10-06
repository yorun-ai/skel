package compiler_test

import (
	"fmt"
	"go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/formatter"
	"go.yorun.ai/skel/internal/parser"
	"strings"
	"testing"
)

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
