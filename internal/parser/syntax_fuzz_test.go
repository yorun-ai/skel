package parser

import "testing"

func FuzzParseSource(f *testing.F) {
	for _, seed := range []string{
		"domain fuzz.demo\n",
		"domain A@A\ndata A{}",
		"domain fuzz.demo\ndata User {\n    id: uuid\n    tags: list<string?>\n}\n",
		"domain fuzz.demo\npub resource File { action read }\nservice Files { method get { require any(File:read) } }\n",
		"domain fuzz.demo\r\ndata User { id: string }\r\n",
		"// only a comment\n",
		"domain fuzz.demo\ndata User {\n  id: string\n",
		"domain fuzz.demo\ndata User {\n  id: \"unterminated\n}\n",
		"domain fuzz.demo\n@desc(\"value\")\ndata User { id: string }\n",
		"",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, source []byte) {
		content, err := ParseSource("fuzz.skel", source)
		if err == nil {
			if content == nil {
				t.Fatal("ParseSource returned no content and no error")
			}
			if validateErr := ValidateSource("fuzz.skel", source); validateErr != nil {
				t.Fatalf("ValidateSource rejected parseable source: %v", validateErr)
			}
		}

		if _, partialErr := ParseSourcePartial("fuzz.skel", source); err == nil && partialErr != nil {
			t.Fatalf("ParseSourcePartial failed for parseable source: %v", partialErr)
		}

		segments := SplitSourceSegments("fuzz.skel", source)
		assertSourceSegmentInvariants(t, source, segments)

		// Recovery parses fragments of the source, so it must survive the same
		// input. Bound the work to keep the fuzz target fast.
		if len(source) > 2048 {
			return
		}
		for index, segment := range segments {
			if index >= 4 {
				return
			}
			if _, fragmentErr := ParseSourceFragment("fuzz.skel", source[segment.Start:segment.End], segment.Line, segment.Start); err == nil && fragmentErr != nil {
				t.Fatalf("ParseSourceFragment failed for a valid segment: %v", fragmentErr)
			}
		}
	})
}
