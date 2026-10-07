package binding

import "testing"

func TestNormalizeTrailingNewline(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "no newline", in: "abc", want: "abc\n"},
		{name: "single newline", in: "abc\n", want: "abc\n"},
		{name: "many newlines", in: "abc\n\n\n", want: "abc\n"},
		{name: "empty", in: "", want: "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTrailingNewline(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeTrailingNewline(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderTemplateReturnsParseAndExecutionErrors(t *testing.T) {
	if _, err := RenderTemplate("{{", nil); err == nil {
		t.Fatal("expected template parse error")
	}
	if _, err := RenderTemplate("{{call .}}", 1); err == nil {
		t.Fatal("expected template execution error")
	}
}
