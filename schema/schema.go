package schema

const (
	// Format identifies schema snapshot JSON produced by skelc.
	Format = "yorun.skel.schema"
	// FormatVersion is the current schema snapshot JSON format version.
	FormatVersion = 1
)

// DeclarationType identifies a top-level Skel declaration in schema JSON.
type DeclarationType string

const (
	// DeclarationTypeActor identifies an actor declaration.
	DeclarationTypeActor DeclarationType = "actor"
	// DeclarationTypeConfig identifies a config declaration.
	DeclarationTypeConfig DeclarationType = "config"
	// DeclarationTypeData identifies a data declaration.
	DeclarationTypeData DeclarationType = "data"
	// DeclarationTypeEnum identifies an enum declaration.
	DeclarationTypeEnum DeclarationType = "enum"
	// DeclarationTypeEvent identifies an event declaration.
	DeclarationTypeEvent DeclarationType = "event"
	// DeclarationTypeResource identifies a resource declaration.
	DeclarationTypeResource DeclarationType = "resource"
	// DeclarationTypeService identifies a service declaration.
	DeclarationTypeService DeclarationType = "service"
	// DeclarationTypeTask identifies a task declaration.
	DeclarationTypeTask DeclarationType = "task"
	// DeclarationTypeWeb identifies a web declaration.
	DeclarationTypeWeb DeclarationType = "web"
)

// ConfigLifecycle identifies the lifetime of a config declaration.
type ConfigLifecycle string

const (
	// ConfigLifecycleEternal identifies an eternal configuration lifecycle.
	ConfigLifecycleEternal ConfigLifecycle = "eternal"
	// ConfigLifecycleInstant identifies an instant configuration lifecycle.
	ConfigLifecycleInstant ConfigLifecycle = "instant"
)

// TypeKind identifies the normalized representation carried by a Type.
type TypeKind string

const (
	// TypeKindScalar identifies a built-in scalar type.
	TypeKindScalar TypeKind = "scalar"
	// TypeKindEnum identifies a normalized enum reference.
	TypeKindEnum TypeKind = "enum"
	// TypeKindData identifies a normalized data reference.
	TypeKindData TypeKind = "data"
	// TypeKindConfig identifies a normalized config reference.
	TypeKindConfig TypeKind = "config"
	// TypeKindEvent identifies a normalized event reference.
	TypeKindEvent TypeKind = "event"
	// TypeKindTypeParameter identifies a generic type parameter.
	TypeKindTypeParameter TypeKind = "typeParameter"
	// TypeKindImportedReference identifies an unresolved imported-domain type.
	TypeKindImportedReference TypeKind = "importedReference"
	// TypeKindList identifies a list type.
	TypeKindList TypeKind = "list"
	// TypeKindMap identifies a map type.
	TypeKindMap TypeKind = "map"
)

// AuthMode identifies the authentication behavior of a service, method, or web.
type AuthMode string

const (
	// AuthModeInherit uses the enclosing service authentication mode on methods.
	AuthModeInherit AuthMode = "inherit"
	// AuthModeUnset inherits authentication behavior from the enclosing context.
	AuthModeUnset AuthMode = "unset"
	// AuthModeRequired requires valid credentials.
	AuthModeRequired AuthMode = "required"
	// AuthModeOptional allows anonymous callers and authenticates supplied credentials.
	AuthModeOptional AuthMode = "optional"
	// AuthModeAnonymous allows only anonymous callers and rejects supplied invalid credentials.
	AuthModeAnonymous AuthMode = "anonymous"
	// AuthModeOff bypasses portal authentication for web declarations.
	AuthModeOff AuthMode = "off"

	// AuthModeAuth requires an authenticated actor.
	AuthModeAuth AuthMode = "auth"
	// AuthModeNoAuth explicitly permits unauthenticated access.
	AuthModeNoAuth AuthMode = "noauth"
)

// RequirementMode identifies one node in a normalized permission expression.
type RequirementMode string

const (
	// RequirementModeCode requires one permission code.
	RequirementModeCode RequirementMode = "code"
	// RequirementModeReference identifies a normalized permission reference.
	RequirementModeReference RequirementMode = "reference"
	// RequirementModeCheck invokes a resource permission check.
	RequirementModeCheck RequirementMode = "check"
	// RequirementModeAll requires every child expression.
	RequirementModeAll RequirementMode = "all"
	// RequirementModeAny requires at least one child expression.
	RequirementModeAny RequirementMode = "any"
)

var declarationKinds = []DeclarationType{
	DeclarationTypeActor, DeclarationTypeConfig, DeclarationTypeData, DeclarationTypeEnum,
	DeclarationTypeEvent, DeclarationTypeResource, DeclarationTypeService, DeclarationTypeTask,
	DeclarationTypeWeb,
}

// Document is the versioned normalized domain emitted by schema snapshot.
type Document struct {
	Format        string         `json:"format"`
	FormatVersion int            `json:"formatVersion"`
	Domain        string         `json:"domain"`
	Description   string         `json:"description,omitempty"`
	Declarations  []*Declaration `json:"declarations"`
}

// Metadata contains common declaration and member documentation metadata.
type Metadata struct {
	Description      string `json:"description,omitempty"`
	Deprecated       bool   `json:"deprecated,omitempty"`
	DeprecatedReason string `json:"deprecatedReason,omitempty"`
}

// Declaration is one complete normalized declaration emitted by schema get or
// contained in a Document.
type Declaration struct {
	Metadata
	Pub      bool            `json:"pub"`
	Name     string          `json:"name"`
	Kind     DeclarationType `json:"type"`
	SkelName string          `json:"skelName"`
	Enum     *EnumSchema     `json:"enum,omitempty"`
	Data     *DataSchema     `json:"data,omitempty"`
	Actor    *ActorSchema    `json:"actor,omitempty"`
	Resource *ResourceSchema `json:"resource,omitempty"`
	Service  *ServiceSchema  `json:"service,omitempty"`
	Web      *WebSchema      `json:"web,omitempty"`
	Task     *TaskSchema     `json:"task,omitempty"`
	Pos      Position        `json:"-"`
}

// Entry is one declaration summary emitted by schema list.
type Entry struct {
	Pub      bool            `json:"pub"`
	Name     string          `json:"name"`
	Kind     DeclarationType `json:"type"`
	SkelName string          `json:"skelName"`
}
