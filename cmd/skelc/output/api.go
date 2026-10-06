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

// SchemaMetadata is part of the schema inspection JSON output.
type SchemaMetadata = internaloutput.SchemaMetadata

// SchemaDeclaration is part of the schema inspection JSON output.
type SchemaDeclaration = internaloutput.SchemaDeclaration

// SchemaEntry is part of the schema inspection JSON output.
type SchemaEntry = internaloutput.SchemaEntry

// SchemaEnum is part of the schema inspection JSON output.
type SchemaEnum = internaloutput.SchemaEnum

// SchemaEnumItem is part of the schema inspection JSON output.
type SchemaEnumItem = internaloutput.SchemaEnumItem

// SchemaData is part of the schema inspection JSON output.
type SchemaData = internaloutput.SchemaData

// SchemaMember is part of the schema inspection JSON output.
type SchemaMember = internaloutput.SchemaMember

// SchemaType is part of the schema inspection JSON output.
type SchemaType = internaloutput.SchemaType

// SchemaActor is part of the schema inspection JSON output.
type SchemaActor = internaloutput.SchemaActor

// SchemaActorAuth is part of the schema inspection JSON output.
type SchemaActorAuth = internaloutput.SchemaActorAuth

// SchemaActorPermission is part of the schema inspection JSON output.
type SchemaActorPermission = internaloutput.SchemaActorPermission

// SchemaActorVia is part of the schema inspection JSON output.
type SchemaActorVia = internaloutput.SchemaActorVia

// SchemaResource is part of the schema inspection JSON output.
type SchemaResource = internaloutput.SchemaResource

// SchemaResourceAction is part of the schema inspection JSON output.
type SchemaResourceAction = internaloutput.SchemaResourceAction

// SchemaResourceCheck is part of the schema inspection JSON output.
type SchemaResourceCheck = internaloutput.SchemaResourceCheck

// SchemaService is part of the schema inspection JSON output.
type SchemaService = internaloutput.SchemaService

// SchemaAudience is part of the schema inspection JSON output.
type SchemaAudience = internaloutput.SchemaAudience

// SchemaMethod is part of the schema inspection JSON output.
type SchemaMethod = internaloutput.SchemaMethod

// SchemaArgument is part of the schema inspection JSON output.
type SchemaArgument = internaloutput.SchemaArgument

// SchemaRequirement is part of the schema inspection JSON output.
type SchemaRequirement = internaloutput.SchemaRequirement

// SchemaRequirementCheck is part of the schema inspection JSON output.
type SchemaRequirementCheck = internaloutput.SchemaRequirementCheck

// SchemaRequirementCheckArgument is part of the schema inspection JSON output.
type SchemaRequirementCheckArgument = internaloutput.SchemaRequirementCheckArgument

// SchemaWeb is part of the schema inspection JSON output.
type SchemaWeb = internaloutput.SchemaWeb

// SchemaTask is part of the schema inspection JSON output.
type SchemaTask = internaloutput.SchemaTask

// SchemaTrigger is part of the schema inspection JSON output.
type SchemaTrigger = internaloutput.SchemaTrigger
