// Package schema implements shared contract types, querying, encoding,
// validation, and compatibility diffing without compiling or loading inputs.
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
	DeclarationTypeActor    DeclarationType = "actor"
	DeclarationTypeConfig   DeclarationType = "config"
	DeclarationTypeData     DeclarationType = "data"
	DeclarationTypeEnum     DeclarationType = "enum"
	DeclarationTypeEvent    DeclarationType = "event"
	DeclarationTypeResource DeclarationType = "resource"
	DeclarationTypeService  DeclarationType = "service"
	DeclarationTypeTask     DeclarationType = "task"
	DeclarationTypeWeb      DeclarationType = "web"
)

// ConfigLifecycle identifies the lifetime of a config declaration.
type ConfigLifecycle string

const (
	ConfigLifecycleEternal ConfigLifecycle = "eternal"
	ConfigLifecycleInstant ConfigLifecycle = "instant"
)

// TypeKind identifies the normalized representation carried by a Type.
type TypeKind string

const (
	TypeKindScalar            TypeKind = "scalar"
	TypeKindEnum              TypeKind = "enum"
	TypeKindData              TypeKind = "data"
	TypeKindConfig            TypeKind = "config"
	TypeKindEvent             TypeKind = "event"
	TypeKindTypeParameter     TypeKind = "typeParameter"
	TypeKindImportedReference TypeKind = "importedReference"
	TypeKindList              TypeKind = "list"
	TypeKindMap               TypeKind = "map"
)

// AuthMode identifies the authentication behavior of a service, method, or web.
type AuthMode string

const (
	// AuthModeInherit uses the enclosing service authentication mode on methods.
	AuthModeInherit AuthMode = "inherit"
	AuthModeUnset   AuthMode = "unset"
	// AuthModeRequired requires valid credentials.
	AuthModeRequired AuthMode = "required"
	// AuthModeOptional allows anonymous callers and authenticates supplied credentials.
	AuthModeOptional AuthMode = "optional"
	// AuthModeAnonymous allows only anonymous callers and rejects supplied invalid credentials.
	AuthModeAnonymous AuthMode = "anonymous"
	// AuthModeOff bypasses portal authentication for web declarations.
	AuthModeOff AuthMode = "off"

	AuthModeAuth   AuthMode = "auth"
	AuthModeNoAuth AuthMode = "noauth"
)

// RequirementMode identifies one node in a normalized permission expression.
type RequirementMode string

const (
	RequirementModeCode      RequirementMode = "code"
	RequirementModeReference RequirementMode = "reference"
	RequirementModeCheck     RequirementMode = "check"
	RequirementModeAll       RequirementMode = "all"
	RequirementModeAny       RequirementMode = "any"
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
