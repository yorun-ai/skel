package source

import (
	"path/filepath"
	"reflect"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/codegen/golang/view"
)

func TestBuildDocGoPayloadUsesDomainDescription(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModelWithDescription("demo.user", "User domain"))
	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})

	payload := gen.buildDocGoPayload()

	if len(payload.CommentLines) == 0 || payload.CommentLines[0] != "Package skeled User domain" {
		t.Fatalf("unexpected doc comment lines: %+v", payload.CommentLines)
	}
}

func TestDeprecatedGoDocLines(t *testing.T) {
	got := deprecatedGoDocLines(nil, "User", "Use Profile instead")
	want := []string{"User", "", "Deprecated: Use Profile instead."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected deprecated docs: got=%v want=%v", got, want)
	}
}

func TestDeprecatedGoDocLinesSupportsMultilineReason(t *testing.T) {
	got := deprecatedGoDocLines(nil, "User", "Use Profile instead\nComplete migration first")
	want := []string{"User", "", "Deprecated: Use Profile instead", "Complete migration first."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected multiline deprecated docs: got=%v want=%v", got, want)
	}
}

func TestBuildDocGoPayloadFallsBackToPackageName(t *testing.T) {
	pkg := buildModelDomainForTest(t, codegentest.DomainModel("demo.user"))
	gen := newGen(Option{
		Domain:      pkg,
		View:        mustView(t, view.ModeFull, pkg),
		Mode:        view.ModeFull,
		PackageName: "skeled",
		Out:         filepath.Join(t.TempDir(), "skeled"),
	})

	payload := gen.buildDocGoPayload()

	if len(payload.CommentLines) == 0 || payload.CommentLines[0] != "Package skeled" {
		t.Fatalf("unexpected fallback doc comment lines: %+v", payload.CommentLines)
	}
}
