package output

import internaloutput "go.yorun.ai/skel/internal/cmd/skelc/output"

const (
	// ExitCodeSuccess identifies a completed result that satisfies the command.
	ExitCodeSuccess = internaloutput.ExitCodeSuccess
	// ExitCodeUnsatisfied identifies a completed check whose result is false.
	ExitCodeUnsatisfied = internaloutput.ExitCodeUnsatisfied
	// ExitCodeError identifies a command that could not produce its normal result.
	ExitCodeError = internaloutput.ExitCodeError

	// ErrorCodeInvalidArgument identifies invalid command arguments.
	ErrorCodeInvalidArgument = internaloutput.ErrorCodeInvalidArgument
	// ErrorCodeCompilationFailed identifies invalid or uncompilable Skel input.
	ErrorCodeCompilationFailed = internaloutput.ErrorCodeCompilationFailed
	// ErrorCodeGitHistoryNotFound identifies an unavailable implicit Git baseline.
	ErrorCodeGitHistoryNotFound = internaloutput.ErrorCodeGitHistoryNotFound
	// ErrorCodeCommandFailed identifies any other command failure.
	ErrorCodeCommandFailed = internaloutput.ErrorCodeCommandFailed
)

// ErrorCode identifies a command failure for programmatic consumers.
type ErrorCode = internaloutput.ErrorCode

// Error is emitted on stdout when a command cannot produce its normal result.
type Error = internaloutput.Error

// CheckResult is emitted by skelc check.
type CheckResult = internaloutput.CheckResult

// FormatResult is emitted by skelc format.
type FormatResult = internaloutput.FormatResult

// GenerationResult is emitted by skelc gen subcommands.
type GenerationResult = internaloutput.GenerationResult

// VersionResult is emitted by skelc version.
type VersionResult = internaloutput.VersionResult

// VersionGolangCodeGenResult reports Vine compatibility for generated Go code.
type VersionGolangCodeGenResult = internaloutput.VersionGolangCodeGenResult

// ScanImport is one direct import returned by skelc scan imports.
type ScanImport = internaloutput.ScanImport
