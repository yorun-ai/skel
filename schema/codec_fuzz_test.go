package schema

import (
	"bytes"
	"testing"
)

func FuzzSchemaDecodeRoundTrip(f *testing.F) {
	var output bytes.Buffer
	if err := Encode(&output, completeTestDocument()); err != nil {
		f.Fatal(err)
	}
	f.Add(output.Bytes())
	f.Add([]byte(`{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[]}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 16384 {
			t.Skip()
		}
		document, err := Decode(bytes.NewReader(input))
		if err != nil {
			return
		}
		var encoded bytes.Buffer
		if err := Encode(&encoded, document); err != nil {
			t.Fatal(err)
		}
		decoded, err := Decode(&encoded)
		if err != nil {
			t.Fatal(err)
		}
		if report, err := Diff(document, decoded); err != nil || len(report.Changes) != 0 {
			t.Fatalf("round trip changed schema: %v, %v", report, err)
		}
	})
}
