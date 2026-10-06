package analyzer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/parser"
)

type cancellationContext struct {
	context.Context
	cancel    context.CancelFunc
	remaining int
}

func (c *cancellationContext) Err() error {
	c.remaining--
	if c.remaining <= 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestAnalysisCancelsInsideLargeDeclaration(t *testing.T) {
	var source strings.Builder
	source.WriteString("domain demo\ndata Large {\n")
	for i := range 1000 {
		fmt.Fprintf(&source, "value%d: string\n", i)
	}
	source.WriteString("}\n")
	content, err := parser.ParseSource("input.skel", []byte(source.String()))
	if err != nil {
		t.Fatal(err)
	}
	for _, imports := range []bool{false, true} {
		t.Run(fmt.Sprint(imports), func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			controlled := &cancellationContext{Context: ctx, cancel: cancel, remaining: 100}
			var analysis *Analysis
			var diagnostics []error
			if imports {
				analysis, diagnostics, err = AnalyzeImportContext(controlled, content)
			} else {
				analysis, diagnostics, err = AnalyzeContext(controlled, content, nil)
			}
			if !errors.Is(err, context.Canceled) || analysis != nil || diagnostics != nil {
				t.Fatalf("cancelled analysis escaped: %v %v %v", analysis, diagnostics, err)
			}
			if _, diagnostics := Analyze(content, nil); len(diagnostics) != 0 {
				t.Fatalf("cancelled work modified source: %v", diagnostics)
			}
		})
	}
}

func TestCompletedAnalysisOutlivesRequestCancellation(t *testing.T) {
	content, err := parser.ParseSource("input.skel", []byte("domain demo\nimport external.demo as external\ndata User {}\n"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	analysis, diagnostics, err := AnalyzeImportContext(ctx, content)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("unexpected result: %v %v", diagnostics, err)
	}
	cancel()
	if len(analysis.ImportNames()) != 1 || analysis.Schema().ReferenceName("external.User") != "external.demo.User" {
		t.Fatal("completed analysis retained the cancelled request context")
	}
}
