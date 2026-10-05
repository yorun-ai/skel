// Package command defines the stable JSON results emitted by skelc commands.
package command

import "go.yorun.ai/skelc/diagnostic"

const (
	ExitCodeSuccess     = 0
	ExitCodeUnsatisfied = 1
	ExitCodeError       = 2
)

// ErrorCode identifies a command failure for programmatic consumers.
type ErrorCode string

const (
	ErrorCodeInvalidArgument    ErrorCode = "INVALID_ARGUMENT"
	ErrorCodeCompilationFailed  ErrorCode = "COMPILATION_FAILED"
	ErrorCodeGitHistoryNotFound ErrorCode = "GIT_HISTORY_NOT_FOUND"
	ErrorCodeCommandFailed      ErrorCode = "COMMAND_FAILED"
)

// Error is emitted on stdout when a command cannot produce its normal result.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e Error) Error() string { return e.Message }

// CheckResult reports whether the input is valid and includes every diagnostic.
type CheckResult struct {
	Valid       bool                    `json:"valid"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
}

// FormatResult reports files changed or requiring formatting.
type FormatResult struct {
	Changed bool     `json:"changed"`
	Files   []string `json:"files"`
}

// GenerationResult reports successful generation completion.
type GenerationResult struct {
	Generated bool `json:"generated"`
}

// VersionResult reports compiler and generated-code compatibility versions.
type VersionResult struct {
	Name          string                     `json:"name"`
	Version       string                     `json:"version"`
	Platform      string                     `json:"platform"`
	GoVersion     string                     `json:"goVersion"`
	GolangCodeGen VersionGolangCodeGenResult `json:"golangCodeGen"`
}

// VersionGolangCodeGenResult reports Vine compatibility for generated Go code.
type VersionGolangCodeGenResult struct {
	MinimumVineVersion string `json:"minimumVineVersion"`
	DefaultVineVersion string `json:"defaultVineVersion"`
}

// ScanImport reports one direct import declaration in the queried input.
// Repeated declarations retain their individual source positions.
type ScanImport struct {
	Domain string `json:"domain"`
	Alias  string `json:"alias,omitempty"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}
