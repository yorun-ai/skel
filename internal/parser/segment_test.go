package parser

import (
	"strings"
	"testing"
)

// assertSourceSegmentInvariants checks the contract every caller of
// SplitSourceSegments relies on: the segments stay inside the source, keep
// declaration order, and cover the source without gaps.
func assertSourceSegmentInvariants(t *testing.T, source []byte, segments []SourceSegment) {
	t.Helper()
	if len(segments) == 0 {
		t.Fatal("expected at least one source segment")
	}
	if segments[0].Start != 0 {
		t.Fatalf("first segment must start at the source start: %+v", segments)
	}
	for index, segment := range segments {
		if segment.Line < 1 {
			t.Fatalf("segment line must be positive: %+v", segment)
		}
		if segment.Start < 0 || segment.End < segment.Start || segment.End > len(source) {
			t.Fatalf("segment range is out of bounds: %+v", segment)
		}
		if index+1 < len(segments) && segment.End != segments[index+1].Start {
			t.Fatalf("segments must be contiguous: %+v", segments)
		}
	}
	if last := segments[len(segments)-1]; last.End != len(source) {
		t.Fatalf("last segment must end at the source end: %+v", segments)
	}
}

func segmentedSources(t *testing.T, source string) []string {
	t.Helper()
	segments := SplitSourceSegments("demo.skel", []byte(source))
	assertSourceSegmentInvariants(t, []byte(source), segments)
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		parts = append(parts, source[segment.Start:segment.End])
	}
	return parts
}

func TestSplitSourceSegmentsKeepsSingleDeclarationTogether(t *testing.T) {
	source := "domain demo.user\ndata User {\n    id: string\n}\n"
	if parts := segmentedSources(t, source); len(parts) != 2 {
		t.Fatalf("expected the domain line and the data declaration: %q", parts)
	}
}

func TestSplitSourceSegmentsSeparatesTopLevelDeclarations(t *testing.T) {
	source := "domain demo.user\n" +
		"data User {\n    id: string\n}\n" +
		"data Order {\n    id: string\n}\n"
	parts := segmentedSources(t, source)
	if len(parts) != 3 {
		t.Fatalf("expected three segments, got %q", parts)
	}
	if !strings.HasPrefix(parts[0], "domain demo.user") {
		t.Fatalf("unexpected first segment: %q", parts[0])
	}
	if !strings.HasPrefix(parts[1], "data User") || !strings.HasPrefix(parts[2], "data Order") {
		t.Fatalf("declarations must start their own segments: %q", parts)
	}
}

func TestSplitSourceSegmentsKeepsDecoratorWithDeclaration(t *testing.T) {
	source := "domain demo.user\n" +
		"data User {\n    id: string\n}\n" +
		"@desc(\"Order payload\")\n" +
		"data Order {\n    id: string\n}\n"
	parts := segmentedSources(t, source)
	if len(parts) != 3 {
		t.Fatalf("expected three segments, got %q", parts)
	}
	if !strings.HasPrefix(parts[2], "@desc(\"Order payload\")") {
		t.Fatalf("decorator must stay with its declaration: %q", parts[2])
	}
}

func TestSplitSourceSegmentsKeepsModifiersWithDeclaration(t *testing.T) {
	source := "domain demo.user\npub data User {\n    id: string\n}\n"
	parts := segmentedSources(t, source)
	if len(parts) != 2 || !strings.HasPrefix(parts[1], "pub data User") {
		t.Fatalf("modifier must stay with the declaration: %q", parts)
	}
}

func TestSplitSourceSegmentsIgnoresCommentsAndNestedBlocks(t *testing.T) {
	source := "domain demo.user\n" +
		"// leading comment\n" +
		"data User {\n" +
		"    // member comment\n" +
		"    label: \"{ not a block }\"\n" +
		"}\n" +
		"/* trailing\n   comment */\n"
	parts := segmentedSources(t, source)
	if len(parts) != 2 {
		t.Fatalf("comments and nested content must not start segments: %q", parts)
	}
	if !strings.Contains(parts[1], "not a block") {
		t.Fatalf("string content must stay inside the declaration: %q", parts[1])
	}
}

func TestSplitSourceSegmentsHandlesCarriageReturns(t *testing.T) {
	source := "domain demo.user\r\ndata User {\r\n    id: string\r\n}\r\n"
	parts := segmentedSources(t, source)
	if len(parts) != 2 || !strings.HasPrefix(parts[1], "data User") {
		t.Fatalf("unexpected CRLF segments: %q", parts)
	}
}

func TestSplitSourceSegmentsKeepsCommentOnlySourceInOneSegment(t *testing.T) {
	source := "// nothing to declare\n"
	parts := segmentedSources(t, source)
	if len(parts) != 1 || parts[0] != source {
		t.Fatalf("unexpected comment-only segments: %q", parts)
	}
}

func TestSplitSourceSegmentsReportsAscendingLines(t *testing.T) {
	source := "domain demo.user\ndata User {\n    id: string\n}\ndata Order {\n    id: string\n}\n"
	segments := SplitSourceSegments("demo.skel", []byte(source))
	assertSourceSegmentInvariants(t, []byte(source), segments)
	for index := 1; index < len(segments); index++ {
		if segments[index].Line <= segments[index-1].Line {
			t.Fatalf("segment lines must ascend: %+v", segments)
		}
	}
}

func TestSplitSourceSegmentsFallsBackToWholeSource(t *testing.T) {
	// Input the lexer rejects must still yield one segment covering the whole
	// source, so callers keep a recoverable range.
	source := "domain demo.user\ndata User {\n    id: string\n}\n\x00"
	segments := SplitSourceSegments("demo.skel", []byte(source))
	assertSourceSegmentInvariants(t, []byte(source), segments)
	if len(segments) != 1 || segments[0].Start != 0 || segments[0].End != len(source) {
		t.Fatalf("lexing failures must fall back to one full segment: %+v", segments)
	}
}

func TestSplitSourceSegmentsKeepsInlineDecoratorWithNextDeclaration(t *testing.T) {
	source := []byte("domain demo @desc(\"next\")\ndata Value {}\ndata Other {}\n")
	segments := SplitSourceSegments("demo.skel", source)
	assertSourceSegmentInvariants(t, source, segments)
	for _, segment := range segments {
		if _, err := ParseSourceFragment("demo.skel", source[segment.Start:segment.End], segment.Line, segment.Start); err != nil {
			t.Fatal(err)
		}
	}
	if len(segments) != 2 {
		t.Fatalf("expected decorated declaration and domain to stay together, got %v", segments)
	}
}
