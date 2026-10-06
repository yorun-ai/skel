package skelc

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
)

type _LogEntry struct {
	Level   string
	Message string
}

type _JSONLEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func errorLogEntry(format string, args ...any) _LogEntry {
	return _LogEntry{Level: logLevelError, Message: fmt.Sprintf(format, args...)}
}

func formatLog(entry _LogEntry, format string) string {
	message := strings.TrimSpace(entry.Message)
	if format == logFormatJSONL {
		return formatLogJSONL(entry.Level, message)
	}
	switch entry.Level {
	case logLevelInfo:
		return "[I] " + stripLogPrefix(message, "[I] ")
	case logLevelWarn:
		return "[W] " + stripLogPrefix(message, "[W] ")
	case logLevelError:
		return "Error: " + message
	default:
		return message
	}
}

func formatLogJSONL(level string, message string) string {
	data, err := json.Marshal(_JSONLEntry{Level: level, Message: normalizeLogMessage(level, message)})
	if err != nil {
		panic(fmt.Sprintf("marshal log entry: %v", err))
	}
	return string(data) + "\n"
}

func normalizeLogMessage(level string, message string) string {
	message = strings.TrimSpace(message)
	switch level {
	case logLevelInfo:
		return stripLogPrefix(message, "[I] ")
	case logLevelWarn:
		return stripLogPrefix(message, "[W] ")
	default:
		return message
	}
}

func stripLogPrefix(message string, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(message), prefix))
}
