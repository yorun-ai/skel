package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	ucli "github.com/urfave/cli/v3"
	commandresult "go.yorun.ai/skelc/internal/command"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/util/logutil"
)

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

const (
	commandSkelc = "skelc"

	ExitCodeSuccess     = commandresult.ExitCodeSuccess
	ExitCodeUnsatisfied = commandresult.ExitCodeUnsatisfied
	ExitCodeError       = commandresult.ExitCodeError

	flagLogFormat = "log-format"
	flagStrict    = "strict"

	logFormatText  = "text"
	logFormatJSONL = "jsonl"

	logLevelWarn  = string(logutil.LevelWarn)
	logLevelError = string(logutil.LevelError)
)

func Main() {
	result := run(os.Args[1:], os.Stdin, os.Stdout)
	if result.Stderr != "" {
		_, _ = fmt.Fprint(os.Stderr, result.Stderr)
		if !strings.HasSuffix(result.Stderr, "\n") {
			_, _ = fmt.Fprintln(os.Stderr)
		}
	}
	os.Exit(result.ExitCode)
}

func Run(args []string) Result {
	var stdout strings.Builder
	result := run(args, strings.NewReader(""), &stdout)
	result.Stdout = stdout.String()
	return result
}

func run(args []string, stdin io.Reader, stdout io.Writer) Result {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return runCLICommand(newCommand(), append([]string{"skelc"}, args...), stdin, stdout)
}

func newCommand() *ucli.Command {
	return &ucli.Command{
		Name:                          commandSkelc,
		Usage:                         "skeleton code generation and checking",
		Suggest:                       true,
		HideHelpCommand:               true,
		CustomHelpTemplate:            groupCommandHelpTemplate,
		CustomRootCommandHelpTemplate: groupCommandHelpTemplate,
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: flagLogFormat, Usage: "log output format: jsonl/text", Value: logFormatJSONL},
			&ucli.BoolFlag{Name: flagStrict, Usage: "reject legacy declarations accepted with migration warnings"},
		},
		Before: func(ctx context.Context, cmd *ucli.Command) (context.Context, error) {
			return ctx, validateLogFormat(cmd)
		},
		Commands: []*ucli.Command{
			newVersionCommand(),
			newLSPCommand(),
			newGenCommand(),
			newSchemaCommand(),
			newScanCommand(),
			newCheckCommand(),
			newFormatCommand(),
		},
	}
}

func runCLICommand(command *ucli.Command, args []string, stdin io.Reader, stdout io.Writer) (result Result) {
	rawLogFormat := rawLogFormatFromArgs(args)
	isJSONCommand := jsonCommandRequested(args)

	var stderr strings.Builder
	var commandStdout strings.Builder

	command.Reader = stdin
	command.Writer = stdout
	if isJSONCommand {
		command.Writer = &commandStdout
	}
	command.ErrWriter = &stderr
	command.ExitErrHandler = func(_ context.Context, _ *ucli.Command, _ error) {}

	err := command.Run(context.Background(), args)
	if err != nil {
		if _, ok := err.(*_CommandUnsatisfied); ok {
			if _, writeErr := io.WriteString(stdout, commandStdout.String()); writeErr != nil {
				return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("write command result: %s", writeErr), rawLogFormat)}
			}
			return Result{ExitCode: ExitCodeUnsatisfied}
		}
		if failure, ok := err.(*_CommandFailure); ok {
			if writeErr := writeJSONTo(stdout, failure.result()); writeErr != nil {
				return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("write command error result: %s", writeErr), rawLogFormat)}
			}
			return Result{ExitCode: ExitCodeError, Stderr: commandFailureLogs(failure, rawLogFormat)}
		}
		if isJSONCommand {
			failure := &commandresult.Error{Code: commandresult.ErrorCodeInvalidArgument, Message: err.Error()}
			if writeErr := writeJSONTo(stdout, failure); writeErr != nil {
				return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("write command error result: %s", writeErr), rawLogFormat)}
			}
			return Result{ExitCode: ExitCodeError}
		}
		if diagnostics, ok := err.(interface{ DiagnosticEntries() compiler.Diagnostics }); ok {
			return Result{
				ExitCode: ExitCodeError,
				Stderr:   formatDiagnostics(diagnostics.DiagnosticEntries(), rawLogFormat),
			}
		}
		if diagnostics, ok := err.(interface{ Errors() []error }); ok {
			return Result{
				ExitCode: ExitCodeError,
				Stderr:   formatErrors(diagnostics.Errors(), rawLogFormat),
			}
		}
		if stderr.Len() > 0 {
			return Result{
				ExitCode: ExitCodeError,
				Stderr:   logutil.Format(logutil.Error("%s", stderr.String()), rawLogFormat),
			}
		}
		return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("%s", err.Error()), rawLogFormat)}
	}
	if stderr.Len() > 0 && !isJSONCommand {
		return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("%s", stderr.String()), rawLogFormat)}
	}
	if isJSONCommand {
		if _, err := io.WriteString(stdout, commandStdout.String()); err != nil {
			return Result{ExitCode: ExitCodeError, Stderr: logutil.Format(logutil.Error("write command result: %s", err), rawLogFormat)}
		}
	}
	return Result{ExitCode: ExitCodeSuccess, Stderr: stderr.String()}
}

func formatDiagnostics(diagnostics compiler.Diagnostics, format string) string {
	if format != logFormatJSONL {
		var output strings.Builder
		for _, diagnostic := range diagnostics {
			level := logutil.LevelError
			if diagnostic.Severity == compiler.DiagnosticSeverityWarning {
				level = logutil.LevelWarn
			}
			formatted := logutil.Format(logutil.Entry{Level: level, Message: diagnostic.Error()}, format)
			output.WriteString(formatted)
			if !strings.HasSuffix(formatted, "\n") {
				output.WriteByte('\n')
			}
		}
		return output.String()
	}
	var output strings.Builder
	for _, diagnostic := range diagnostics {
		level := logLevelError
		if diagnostic.Severity == compiler.DiagnosticSeverityWarning {
			level = logLevelWarn
		}
		entry := struct {
			Level string `json:"level"`
			compiler.Diagnostic
		}{Level: level, Diagnostic: diagnostic}
		content, err := json.Marshal(entry)
		if err != nil {
			return formatErrors([]error{err}, format)
		}
		output.Write(content)
		output.WriteByte('\n')
	}
	return output.String()
}

func formatErrors(errors []error, format string) string {
	var output strings.Builder
	for _, err := range errors {
		formatted := logutil.Format(logutil.Entry{Level: logutil.LevelError, Message: err.Error()}, format)
		output.WriteString(formatted)
		if !strings.HasSuffix(formatted, "\n") {
			output.WriteByte('\n')
		}
	}
	return output.String()
}

func writeWarningLogs(cmd *ucli.Command, diagnostics compiler.Diagnostics) {
	for _, item := range diagnostics {
		if item.Severity != compiler.DiagnosticSeverityWarning {
			continue
		}
		_, _ = fmt.Fprint(cmd.Root().ErrWriter, formatDiagnostics(compiler.Diagnostics{item}, commandLogFormat(cmd)))
	}
}

func commandLogFormat(cmd *ucli.Command) string {
	value := cmd.String(flagLogFormat)
	if value == "" {
		value = cmd.Root().String(flagLogFormat)
	}
	if value == "" {
		return logFormatJSONL
	}
	return value
}

func normalizeLogFormat(value string) (string, error) {
	if value == "" {
		return logFormatJSONL, nil
	}
	if value != logFormatText && value != logFormatJSONL {
		return "", fmt.Errorf("invalid log-format %q, expected jsonl/text", value)
	}
	return value, nil
}

func validateLogFormat(cmd *ucli.Command) error {
	_, err := normalizeLogFormat(commandLogFormat(cmd))
	return err
}

func rawLogFormatFromArgs(args []string) string {
	for index, arg := range args {
		if arg == "--"+flagLogFormat && index+1 < len(args) {
			return args[index+1]
		}
		if value, ok := strings.CutPrefix(arg, "--"+flagLogFormat+"="); ok {
			return value
		}
	}
	return logFormatJSONL
}

func parseMappingFlags(values []string, flagName string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	parsed := map[string]string{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok || key == "" || val == "" {
			return nil, fmt.Errorf("invalid flag %s value %q, expected name=value", flagName, value)
		}
		_, duplicated := parsed[key]
		if duplicated {
			return nil, fmt.Errorf("duplicated flag %s key %q", flagName, key)
		}
		parsed[key] = val
	}
	return parsed, nil
}
