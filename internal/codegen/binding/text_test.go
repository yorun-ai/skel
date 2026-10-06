package binding

import (
	"reflect"
	"testing"
)

func TestSplitDocLines(t *testing.T) {
	lines := SplitDocLines(" First line \n  Second line  \n\n Third line ")
	want := []string{"First line", "Second line", "", "Third line"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("unexpected lines: got=%v want=%v", lines, want)
	}
}

func TestSplitDocLinesReturnsNilForBlankInput(t *testing.T) {
	if lines := SplitDocLines(" \n\t "); lines != nil {
		t.Fatalf("expected nil lines, got=%v", lines)
	}
}

func TestMergeDescriptionAndExample(t *testing.T) {
	got := MergeDescriptionAndExample("Pagination cursor", `"cursor-1"`)
	want := `Pagination cursor (e.g. "cursor-1")`
	if got != want {
		t.Fatalf("unexpected merged comment: got=%q want=%q", got, want)
	}
}

func TestMergeDescriptionAndExampleWithoutDescription(t *testing.T) {
	got := MergeDescriptionAndExample("", `{ id:10001, name:"zhangsan" }`)
	want := `e.g. { id:10001, name:"zhangsan" }`
	if got != want {
		t.Fatalf("unexpected merged comment without description: got=%q want=%q", got, want)
	}
}

func TestCompactDocValue(t *testing.T) {
	got := CompactDocValue("{\n  id:10001,\n  name:\"zhangsan\",\n  email:\"a@b.com\"\n}")
	want := `{ id:10001, name:"zhangsan", email:"a@b.com" }`
	if got != want {
		t.Fatalf("unexpected compact doc value: got=%q want=%q", got, want)
	}
}

func TestCompactDocValuePreservesSpacesInsideString(t *testing.T) {
	got := CompactDocValue(`{ label:"hello world", note:"a  b" }`)
	want := `{ label:"hello world", note:"a  b" }`
	if got != want {
		t.Fatalf("unexpected compact doc value with string spaces: got=%q want=%q", got, want)
	}
}
