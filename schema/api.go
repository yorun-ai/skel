package schema

import (
	"io"

	internalschema "go.yorun.ai/skelc/internal/schema"
)

const (
	// Format identifies schema snapshot JSON produced by skelc.
	Format = internalschema.Format
	// FormatVersion is the current schema snapshot JSON format version.
	FormatVersion = internalschema.FormatVersion

	// DeclarationTypeActor identifies an actor declaration.
	DeclarationTypeActor = internalschema.DeclarationTypeActor
	// DeclarationTypeConfig identifies a config declaration.
	DeclarationTypeConfig = internalschema.DeclarationTypeConfig
	// DeclarationTypeData identifies a data declaration.
	DeclarationTypeData = internalschema.DeclarationTypeData
	// DeclarationTypeEnum identifies an enum declaration.
	DeclarationTypeEnum = internalschema.DeclarationTypeEnum
	// DeclarationTypeEvent identifies an event declaration.
	DeclarationTypeEvent = internalschema.DeclarationTypeEvent
	// DeclarationTypeResource identifies a resource declaration.
	DeclarationTypeResource = internalschema.DeclarationTypeResource
	// DeclarationTypeService identifies a service declaration.
	DeclarationTypeService = internalschema.DeclarationTypeService
	// DeclarationTypeTask identifies a task declaration.
	DeclarationTypeTask = internalschema.DeclarationTypeTask
	// DeclarationTypeWeb identifies a web declaration.
	DeclarationTypeWeb = internalschema.DeclarationTypeWeb

	// ConfigLifecycleEternal identifies an eternal configuration lifecycle.
	ConfigLifecycleEternal = internalschema.ConfigLifecycleEternal
	// ConfigLifecycleInstant identifies an instant configuration lifecycle.
	ConfigLifecycleInstant = internalschema.ConfigLifecycleInstant

	// TypeKindScalar identifies a built-in scalar type.
	TypeKindScalar = internalschema.TypeKindScalar
	// TypeKindEnum identifies a normalized enum reference.
	TypeKindEnum = internalschema.TypeKindEnum
	// TypeKindData identifies a normalized data reference.
	TypeKindData = internalschema.TypeKindData
	// TypeKindConfig identifies a normalized config reference.
	TypeKindConfig = internalschema.TypeKindConfig
	// TypeKindEvent identifies a normalized event reference.
	TypeKindEvent = internalschema.TypeKindEvent
	// TypeKindTypeParameter identifies a generic type parameter.
	TypeKindTypeParameter = internalschema.TypeKindTypeParameter
	// TypeKindImportedReference identifies an unresolved imported-domain type.
	TypeKindImportedReference = internalschema.TypeKindImportedReference
	// TypeKindList identifies a list type.
	TypeKindList = internalschema.TypeKindList
	// TypeKindMap identifies a map type.
	TypeKindMap = internalschema.TypeKindMap

	// AuthModeInherit uses the enclosing service authentication mode on methods.
	AuthModeInherit = internalschema.AuthModeInherit
	// AuthModeUnset inherits authentication behavior from the enclosing context.
	AuthModeUnset = internalschema.AuthModeUnset
	// AuthModeRequired requires valid credentials.
	AuthModeRequired = internalschema.AuthModeRequired
	// AuthModeOptional allows anonymous callers and authenticates supplied credentials.
	AuthModeOptional = internalschema.AuthModeOptional
	// AuthModeAnonymous allows only anonymous callers and rejects supplied invalid credentials.
	AuthModeAnonymous = internalschema.AuthModeAnonymous
	// AuthModeOff bypasses portal authentication for web declarations.
	AuthModeOff = internalschema.AuthModeOff

	// AuthModeAuth requires an authenticated actor.
	AuthModeAuth = internalschema.AuthModeAuth
	// AuthModeNoAuth explicitly permits unauthenticated access.
	AuthModeNoAuth = internalschema.AuthModeNoAuth

	// RequirementModeCode requires one permission code.
	RequirementModeCode = internalschema.RequirementModeCode
	// RequirementModeReference identifies a normalized permission reference.
	RequirementModeReference = internalschema.RequirementModeReference
	// RequirementModeCheck invokes a resource permission check.
	RequirementModeCheck = internalschema.RequirementModeCheck
	// RequirementModeAll requires every child expression.
	RequirementModeAll = internalschema.RequirementModeAll
	// RequirementModeAny requires at least one child expression.
	RequirementModeAny = internalschema.RequirementModeAny
)

// DeclarationType identifies a top-level Skel declaration in schema JSON.
type DeclarationType = internalschema.DeclarationType

// ConfigLifecycle identifies the lifetime of a config declaration.
type ConfigLifecycle = internalschema.ConfigLifecycle

// TypeKind identifies the normalized representation carried by a Type.
type TypeKind = internalschema.TypeKind

// AuthMode identifies the authentication behavior of a service, method, or web.
type AuthMode = internalschema.AuthMode

// RequirementMode identifies one node in a normalized permission expression.
type RequirementMode = internalschema.RequirementMode

// Document is the versioned normalized domain emitted by schema snapshot.
type Document = internalschema.Document

// Metadata contains common declaration and member documentation metadata.
type Metadata = internalschema.Metadata

// Declaration is one complete normalized declaration emitted by schema get.
type Declaration = internalschema.Declaration

// EnumSchema is the normalized body of an enum declaration.
type EnumSchema = internalschema.EnumSchema

// EnumItem is one normalized enum item.
type EnumItem = internalschema.EnumItem

// DataSchema is the normalized body shared by data, config, and event declarations.
type DataSchema = internalschema.DataSchema

// Member is one normalized structured-data member.
type Member = internalschema.Member

// Type is a normalized type expression.
type Type = internalschema.Type

// ActorSchema is the normalized body of an actor declaration.
type ActorSchema = internalschema.ActorSchema

// ActorVia is one actor transport capability.
type ActorVia = internalschema.ActorVia

// ResourceSchema is the normalized body of a resource declaration.
type ResourceSchema = internalschema.ResourceSchema

// ResourceAction is one normalized resource action.
type ResourceAction = internalschema.ResourceAction

// ResourceCheck is one normalized resource permission check.
type ResourceCheck = internalschema.ResourceCheck

// ServiceSchema is the normalized body of a service declaration.
type ServiceSchema = internalschema.ServiceSchema

// Audience is one normalized actor audience.
type Audience = internalschema.Audience

// Method is one normalized service method.
type Method = internalschema.Method

// Argument is one normalized method, check, or trigger argument.
type Argument = internalschema.Argument

// Requirement is one node in a normalized permission expression.
type Requirement = internalschema.Requirement

// RequirementCheck is one normalized permission check invocation.
type RequirementCheck = internalschema.RequirementCheck

// RequirementCheckArgument binds one permission-check argument.
type RequirementCheckArgument = internalschema.RequirementCheckArgument

// WebSchema is the normalized body of a web declaration.
type WebSchema = internalschema.WebSchema

// TaskSchema is the normalized body of a task declaration.
type TaskSchema = internalschema.TaskSchema

// Trigger is one normalized task trigger.
type Trigger = internalschema.Trigger

// Entry is one declaration summary emitted by schema list.
type Entry = internalschema.Entry

const (
	// ImpactBreaking identifies a structurally incompatible change.
	ImpactBreaking = internalschema.ImpactBreaking
	// ImpactDangerous identifies a structurally compatible semantic change.
	ImpactDangerous = internalschema.ImpactDangerous
	// ImpactCompatible identifies a compatible change.
	ImpactCompatible = internalschema.ImpactCompatible

	// ChangeAdded identifies an added schema element.
	ChangeAdded = internalschema.ChangeAdded
	// ChangeRemoved identifies a removed schema element.
	ChangeRemoved = internalschema.ChangeRemoved
	// ChangeModified identifies a modified schema element.
	ChangeModified = internalschema.ChangeModified
)

// ImpactLevel classifies a schema change's compatibility impact.
type ImpactLevel = internalschema.ImpactLevel

// ChangeType identifies whether a schema element was added, removed, or modified.
type ChangeType = internalschema.ChangeType

// Change is one change in a Report.
type Change = internalschema.Change

// Summary contains schema diff counts grouped by impact level.
type Summary = internalschema.Summary

// Report is the complete JSON report emitted by schema diff.
type Report = internalschema.Report

// Encode validates and writes one indented schema snapshot JSON document.
func Encode(writer io.Writer, document *Document) error {
	return internalschema.Encode(writer, document)
}

// Decode reads exactly one schema snapshot JSON document, rejects unknown
// fields, and validates its format version and normalized structure.
func Decode(reader io.Reader) (*Document, error) {
	return internalschema.Decode(reader)
}

// Validate checks a schema snapshot's format version and normalized structure.
func Validate(document *Document) error {
	return internalschema.Validate(document)
}
