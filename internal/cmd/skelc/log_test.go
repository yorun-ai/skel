package skelc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorLogEntry(t *testing.T) {
	entry := errorLogEntry("failed to %s the contract", "compile")
	if entry.Level != logLevelError {
		t.Fatalf("unexpected level: %q", entry.Level)
	}
	if entry.Message != "failed to compile the contract" {
		t.Fatalf("unexpected message: %q", entry.Message)
	}
}

func TestFormatLogRendersTextLevels(t *testing.T) {
	for _, test := range []struct {
		name  string
		entry _LogEntry
		want  string
	}{
		{name: "info", entry: _LogEntry{Level: logLevelInfo, Message: "  [I] ready  "}, want: "[I] ready"},
		{name: "info without prefix", entry: _LogEntry{Level: logLevelInfo, Message: "ready"}, want: "[I] ready"},
		{name: "warn", entry: _LogEntry{Level: logLevelWarn, Message: "[W] slow"}, want: "[W] slow"},
		{name: "error keeps text", entry: _LogEntry{Level: logLevelError, Message: "  boom  "}, want: "Error: boom"},
		{name: "unknown level", entry: _LogEntry{Level: "trace", Message: "  detail  "}, want: "detail"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := formatLog(test.entry, logFormatText); got != test.want {
				t.Fatalf("formatLog() = %q; want %q", got, test.want)
			}
		})
	}
}

func TestFormatLogRendersJSONL(t *testing.T) {
	got := formatLog(_LogEntry{Level: logLevelInfo, Message: "  [I] ready  "}, logFormatJSONL)
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("JSONL entries must end with a newline: %q", got)
	}

	var entry _JSONLEntry
	if err := json.Unmarshal([]byte(got), &entry); err != nil {
		t.Fatalf("decode JSONL entry: %v", err)
	}
	if entry.Level != "info" || entry.Message != "ready" {
		t.Fatalf("unexpected JSONL entry: %+v", entry)
	}
}

func TestFormatLogKeepsErrorTextInJSONL(t *testing.T) {
	got := formatLog(_LogEntry{Level: logLevelError, Message: "  Error: boom  "}, logFormatJSONL)

	var entry _JSONLEntry
	if err := json.Unmarshal([]byte(got), &entry); err != nil {
		t.Fatalf("decode JSONL entry: %v", err)
	}
	if entry.Level != "error" || entry.Message != "Error: boom" {
		t.Fatalf("unexpected JSONL entry: %+v", entry)
	}
}
