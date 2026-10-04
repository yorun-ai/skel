package model

import internalmodel "go.yorun.ai/skelc/internal/model"

const (
	// ActorViaClient represents a client application transport.
	ActorViaClient = internalmodel.ActorViaClient
	// ActorViaAgent represents an agent transport.
	ActorViaAgent = internalmodel.ActorViaAgent
	// ActorViaOpenAPI represents an OpenAPI transport.
	ActorViaOpenAPI = internalmodel.ActorViaOpenAPI

	// DataKindData identifies a data declaration.
	DataKindData = internalmodel.DataKindData
	// DataKindConfig identifies a config declaration.
	DataKindConfig = internalmodel.DataKindConfig
	// DataKindEvent identifies an event payload declaration.
	DataKindEvent = internalmodel.DataKindEvent

	// ConfigLifecycleEternal identifies the eternal config lifecycle.
	ConfigLifecycleEternal = internalmodel.ConfigLifecycleEternal
	// ConfigLifecycleInstant identifies the instant config lifecycle.
	ConfigLifecycleInstant = internalmodel.ConfigLifecycleInstant

	// AuthModeUnset inherits authentication behavior from the enclosing context.
	AuthModeUnset = internalmodel.AuthModeUnset
	// AuthModeRequired requires valid credentials.
	AuthModeRequired = internalmodel.AuthModeRequired
	// AuthModeOptional allows anonymous callers and authenticates supplied credentials.
	AuthModeOptional = internalmodel.AuthModeOptional
	// AuthModeGuest allows only anonymous callers and rejects supplied invalid credentials.
	AuthModeGuest = internalmodel.AuthModeGuest
	// AuthModeOff bypasses portal authentication for web declarations.
	AuthModeOff = internalmodel.AuthModeOff

	// AuthModeAuth requires an authenticated actor.
	AuthModeAuth = internalmodel.AuthModeAuth
	// AuthModeNoAuth explicitly allows unauthenticated access.
	AuthModeNoAuth = internalmodel.AuthModeNoAuth

	// PermissionRequireModeCode requires one resource action code.
	PermissionRequireModeCode = internalmodel.PermissionRequireModeCode
	// PermissionRequireModeCheck invokes a resource permission check.
	PermissionRequireModeCheck = internalmodel.PermissionRequireModeCheck
	// PermissionRequireModeAll requires every child expression to pass.
	PermissionRequireModeAll = internalmodel.PermissionRequireModeAll
	// PermissionRequireModeAny requires at least one child expression to pass.
	PermissionRequireModeAny = internalmodel.PermissionRequireModeAny

	// TypeKindUnresolvedReference identifies a named reference that has not yet
	// been resolved during semantic analysis.
	TypeKindUnresolvedReference = internalmodel.TypeKindUnresolvedReference
	// TypeKindScalar identifies a built-in scalar type.
	TypeKindScalar = internalmodel.TypeKindScalar
	// TypeKindList identifies a list type.
	TypeKindList = internalmodel.TypeKindList
	// TypeKindMap identifies a map type.
	TypeKindMap = internalmodel.TypeKindMap
	// TypeKindEnum identifies a resolved enum reference.
	TypeKindEnum = internalmodel.TypeKindEnum
	// TypeKindData identifies a resolved data reference.
	TypeKindData = internalmodel.TypeKindData
	// TypeKindTypeParameter identifies a generic type parameter reference.
	TypeKindTypeParameter = internalmodel.TypeKindTypeParameter

	// ScalarInt identifies an integer.
	ScalarInt = internalmodel.ScalarInt
	// ScalarFloat identifies a floating-point number.
	ScalarFloat = internalmodel.ScalarFloat
	// ScalarBoolean identifies a boolean.
	ScalarBoolean = internalmodel.ScalarBoolean
	// ScalarString identifies a UTF-8 string.
	ScalarString = internalmodel.ScalarString
	// ScalarDecimal identifies an exact decimal value.
	ScalarDecimal = internalmodel.ScalarDecimal
	// ScalarBinary identifies binary data.
	ScalarBinary = internalmodel.ScalarBinary
	// ScalarTimestamp identifies an absolute timestamp.
	ScalarTimestamp = internalmodel.ScalarTimestamp
	// ScalarDuration identifies a duration.
	ScalarDuration = internalmodel.ScalarDuration
	// ScalarLocalDate identifies a calendar date without a timezone.
	ScalarLocalDate = internalmodel.ScalarLocalDate
	// ScalarLocalTime identifies a time of day without a timezone.
	ScalarLocalTime = internalmodel.ScalarLocalTime
	// ScalarLocalDateTime identifies a date and time without a timezone.
	ScalarLocalDateTime = internalmodel.ScalarLocalDateTime
	// ScalarUUID identifies a UUID.
	ScalarUUID = internalmodel.ScalarUUID
	// ScalarJSON identifies an arbitrary JSON value.
	ScalarJSON = internalmodel.ScalarJSON
)

// ActorViaKind identifies a transport through which an actor can access a domain.
type ActorViaKind = internalmodel.ActorViaKind

// Actor describes a caller identity and its authorization facilities.
type Actor = internalmodel.Actor

// ActorVia describes one transport declared by an actor.
type ActorVia = internalmodel.ActorVia

// DataKind identifies the Skel declaration represented by a Data value.
type DataKind = internalmodel.DataKind

// ConfigLifecycle controls the lifetime of generated configuration values.
type ConfigLifecycle = internalmodel.ConfigLifecycle

// Data describes a data, config, or event payload declaration.
type Data = internalmodel.Data

// DataMember describes one member of a Data declaration.
type DataMember = internalmodel.DataMember

// Domain is the validated semantic model for one Skel domain.
type Domain = internalmodel.Domain

// DomainSpec contains the values used to construct a Domain.
type DomainSpec = internalmodel.DomainSpec

// Import describes one domain imported by a Skel contract.
type Import = internalmodel.Import

// Enum describes an enum declaration.
type Enum = internalmodel.Enum

// EnumItem describes one item in an enum declaration.
type EnumItem = internalmodel.EnumItem

// PermissionRequireMode identifies the form of a permission expression.
type PermissionRequireMode = internalmodel.PermissionRequireMode

// PermissionRequire describes a normalized require clause.
type PermissionRequire = internalmodel.PermissionRequire

// PermissionExpr is one node in a normalized permission expression tree.
type PermissionExpr = internalmodel.PermissionExpr

// PermissionCheckInvocation describes a resolved permission-check method call.
type PermissionCheckInvocation = internalmodel.PermissionCheckInvocation

// PermissionCheckArgument describes one resolved permission-check argument.
type PermissionCheckArgument = internalmodel.PermissionCheckArgument

// Position identifies a one-based location in a Skel source file.
type Position = internalmodel.Position

// Resource describes a permission resource and its actions and checks.
type Resource = internalmodel.Resource

// ResourceAction describes one permission-bearing action on a resource.
type ResourceAction = internalmodel.ResourceAction

// ResourceCheck describes a named permission check and its generated method.
type ResourceCheck = internalmodel.ResourceCheck

// AuthMode controls whether authentication is required for a service or method.
type AuthMode = internalmodel.AuthMode

// Service describes a callable service declaration.
type Service = internalmodel.Service

// ActorAudience identifies an actor and transport allowed to access an entry point.
type ActorAudience = internalmodel.ActorAudience

// Method describes one callable service method.
type Method = internalmodel.Method

// Argument describes one service-method or task-trigger argument.
type Argument = internalmodel.Argument

// ArgumentSource identifies who supplies an argument's value.
type ArgumentSource = internalmodel.ArgumentSource

const (
	// ArgumentSourceDeclared identifies an argument supplied by the caller.
	ArgumentSourceDeclared = internalmodel.ArgumentSourceDeclared
	// ArgumentSourcePermissionCode identifies a runtime-injected permission code.
	ArgumentSourcePermissionCode = internalmodel.ArgumentSourcePermissionCode
)

// Task describes a background task declaration.
type Task = internalmodel.Task

// TaskTrigger describes one way to invoke a task.
type TaskTrigger = internalmodel.TaskTrigger

// TypeParameter describes a generic data type parameter.
type TypeParameter = internalmodel.TypeParameter

// Type describes a semantic type.
type Type = internalmodel.Type

// TypeKind identifies the representation carried by a Type.
type TypeKind = internalmodel.TypeKind

// Scalar identifies a built-in Skel scalar type.
type Scalar = internalmodel.Scalar

// ListType describes the element type of a list.
type ListType = internalmodel.ListType

// MapType describes the key and value types of a map.
type MapType = internalmodel.MapType

// Web describes a web entry point and the actors allowed to access it.
type Web = internalmodel.Web

// NewDomainFromSpec constructs a domain from already validated semantic data.
// It does not validate, normalize, copy, or hash the supplied values.
func NewDomainFromSpec(spec DomainSpec) *Domain {
	return internalmodel.NewDomainFromSpec(spec)
}
