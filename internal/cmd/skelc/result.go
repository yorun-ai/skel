package skelc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
	compiler "go.yorun.ai/skel/internal/compiler"
)

func writeJSONResult(cmd *ucli.Command, value any, context string) error {
	if err := writeJSONTo(cmd.Root().Writer, value); err != nil {
		return fmt.Errorf("encode %s: %w", context, err)
	}
	return nil
}

func writeJSONTo(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

type _CommandFailure struct {
	code  output.ErrorCode
	cause error
}

func (e *_CommandFailure) Error() string { return e.cause.Error() }
func (e *_CommandFailure) Unwrap() error { return e.cause }

func (e *_CommandFailure) result() *output.Error {
	return &output.Error{Code: e.code, Message: e.cause.Error()}
}

func commandFailure(code output.ErrorCode, err error) error {
	if failure, ok := err.(*_CommandFailure); ok {
		return failure
	}
	return &_CommandFailure{code: code, cause: err}
}

type _CommandUnsatisfied struct{}

func (*_CommandUnsatisfied) Error() string { return "command result is unsatisfied" }

func commandUnsatisfied() error { return new(_CommandUnsatisfied) }

func commandFailureLogs(failure *_CommandFailure, format string) string {
	var diagnosticEntries interface{ DiagnosticEntries() compiler.Diagnostics }
	if errors.As(failure.cause, &diagnosticEntries) {
		return formatDiagnostics(diagnosticEntries.DiagnosticEntries(), format)
	}
	var diagnosticErrors interface{ Errors() []error }
	if errors.As(failure.cause, &diagnosticErrors) {
		return formatErrors(diagnosticErrors.Errors(), format)
	}
	return ""
}

func jsonCommandRequested(args []string) bool {
	for index := 1; index < len(args); index++ {
		arg := args[index]
		if arg == "--"+flagLogFormat {
			index++
			continue
		}
		if strings.HasPrefix(arg, "--"+flagLogFormat+"=") {
			continue
		}
		if arg == "--"+flagStrict || strings.HasPrefix(arg, "--"+flagStrict+"=") {
			continue
		}
		if arg == "--help" || arg == "-h" || arg == commandLSP {
			return false
		}
		return true
	}
	return false
}
