package skelc

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestRunSkelcLSPUsesRegisteredCommand(t *testing.T) {
	original := serveLSP
	t.Cleanup(func() { serveLSP = original })

	var called bool
	serveLSP = func(_ context.Context, input io.Reader, output io.Writer, strict bool) error {
		called = true
		if !strict {
			t.Fatal("strict mode was not passed to LSP")
		}
		request, err := io.ReadAll(input)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "response:%s", request)
		return err
	}

	var stdout strings.Builder
	result := run([]string{"--log-format", "text", "--strict", "lsp"}, strings.NewReader("request"), &stdout)

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	if !called {
		t.Fatal("expected the registered lsp command to start the server")
	}
	if stdout.String() != "response:request" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
	if result.Stderr != "" {
		t.Fatalf("unexpected stderr: %q", result.Stderr)
	}
}

func TestRunSkelcLSPHelpDoesNotStartServer(t *testing.T) {
	original := serveLSP
	t.Cleanup(func() { serveLSP = original })

	serveLSP = func(context.Context, io.Reader, io.Writer, bool) error {
		t.Fatal("lsp server started while rendering help")
		return nil
	}

	result := Run([]string{"lsp", "--help"})

	if result.ExitCode != ExitCodeSuccess {
		t.Fatalf("unexpected exit code: %d, stderr=%q", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "skelc lsp") {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("unexpected stderr: %q", result.Stderr)
	}
}

func TestRunSkelcLSPReportsServerError(t *testing.T) {
	original := serveLSP
	t.Cleanup(func() { serveLSP = original })

	serveLSP = func(context.Context, io.Reader, io.Writer, bool) error {
		return fmt.Errorf("serve lsp")
	}

	result := Run([]string{"lsp"})

	if result.ExitCode != ExitCodeError {
		t.Fatalf("expected error exit code, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Stderr, "serve lsp") {
		t.Fatalf("unexpected stderr: %q", result.Stderr)
	}
}
