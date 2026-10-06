package skelc

import (
	"bytes"
	"errors"
	"testing"

	ucli "github.com/urfave/cli/v3"
)

type _ErrorWriter struct {
	err error
}

func (w _ErrorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestWriteJSONResultUsesStableEncoding(t *testing.T) {
	var output bytes.Buffer
	cmd := &ucli.Command{Writer: &output}
	if err := writeJSONResult(cmd, map[string]string{"value": "<>&"}, "test value"); err != nil {
		t.Fatal(err)
	}
	if output.String() != "{\n  \"value\": \"<>&\"\n}\n" {
		t.Fatalf("unexpected JSON output:\n%s", output.String())
	}
}

func TestWriteJSONResultReturnsWriterErrors(t *testing.T) {
	want := errors.New("write failed")
	cmd := &ucli.Command{Writer: _ErrorWriter{err: want}}
	err := writeJSONResult(cmd, map[string]string{"value": "test"}, "test value")
	if !errors.Is(err, want) {
		t.Fatalf("expected writer error, got %v", err)
	}
}
