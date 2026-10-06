package cli

import internalcommand "go.yorun.ai/skel/internal/command"

const (
	// ExitCodeSuccess identifies a completed result that satisfies the command.
	ExitCodeSuccess = internalcommand.ExitCodeSuccess
	// ExitCodeUnsatisfied identifies a completed check whose result is false.
	ExitCodeUnsatisfied = internalcommand.ExitCodeUnsatisfied
	// ExitCodeError identifies a command that could not produce its normal result.
	ExitCodeError = internalcommand.ExitCodeError

	// ErrorCodeInvalidArgument identifies invalid command arguments.
	ErrorCodeInvalidArgument = internalcommand.ErrorCodeInvalidArgument
	// ErrorCodeCompilationFailed identifies invalid or uncompilable Skel input.
	ErrorCodeCompilationFailed = internalcommand.ErrorCodeCompilationFailed
	// ErrorCodeGitHistoryNotFound identifies an unavailable implicit Git baseline.
	ErrorCodeGitHistoryNotFound = internalcommand.ErrorCodeGitHistoryNotFound
	// ErrorCodeCommandFailed identifies any other command failure.
	ErrorCodeCommandFailed = internalcommand.ErrorCodeCommandFailed
)

// ErrorCode identifies a command failure for programmatic consumers.
type ErrorCode = internalcommand.ErrorCode

// Error is emitted on stdout when a command cannot produce its normal result.
type Error = internalcommand.Error

// CheckResult is emitted by skelc check.
type CheckResult = internalcommand.CheckResult

// FormatResult is emitted by skelc format.
type FormatResult = internalcommand.FormatResult

// GenerationResult is emitted by skelc gen subcommands.
type GenerationResult = internalcommand.GenerationResult

// VersionResult is emitted by skelc version.
type VersionResult = internalcommand.VersionResult

// VersionGolangCodeGenResult reports Vine compatibility for generated Go code.
type VersionGolangCodeGenResult = internalcommand.VersionGolangCodeGenResult

// ScanImport is one direct import returned by skelc scan imports.
type ScanImport = internalcommand.ScanImport

// SchemaImport retains the existing Go API name for an import declaration.
// Deprecated: use ScanImport.
type SchemaImport = ScanImport
