// Package optionvalidation defines the validation contract shared by the public API and
// CLI adapters without depending on either package.
package optionvalidation

// Field identifies one normalized generation option.
type Field string

const (
	FieldApiActor                 Field = "api.actor"
	FieldApiPrune                 Field = "api.prune"
	FieldApiType                  Field = "api.type"
	FieldSkelInput                Field = "skel.input"
	FieldSkelImport               Field = "skel.import"
	FieldGoOutput                 Field = "go.output"
	FieldGoVineVersion            Field = "go.vine-version"
	FieldGoVrpcVersion            Field = "go.vrpc-version"
	FieldGoCompilerVersion        Field = "go.compiler-version"
	FieldGoModuleIdentity         Field = "go.module-identity"
	FieldGoPublicOutput           Field = "go.public-output"
	FieldGoPublicModule           Field = "go.public-module"
	FieldGoModule                 Field = "go.module"
	FieldGoModulePrefix           Field = "go.module-prefix"
	FieldGoImport                 Field = "go.import"
	FieldTypeScriptOutput         Field = "typescript.output"
	FieldTypeScriptImport         Field = "typescript.import"
	FieldTypeScriptModuleIdentity Field = "typescript.module-identity"
	FieldTypeScriptModule         Field = "typescript.module"
	FieldTypeScriptModuleScope    Field = "typescript.module-scope"
	FieldSkeletonOutput           Field = "skeleton.output"
	FieldSkeletonPublicOnly       Field = "skeleton.public-only"
)

// Rule identifies the validation constraint that a field violated.
type Rule string

const (
	RuleRequiresApi          Rule = "requires-api"
	RuleRequiresPrune        Rule = "requires-prune"
	RuleRequired             Rule = "required"
	RuleRequiresModule       Rule = "requires-module"
	RuleRequiresPublicOutput Rule = "requires-public-output"
	RuleNoTrailingSlash      Rule = "no-trailing-slash"
	RuleInvalid              Rule = "invalid"
)

// ValidationError carries a typed field/rule pair and an API-facing message.
type ValidationError struct {
	Field   Field
	Rule    Rule
	Message string
}

// NewValidationError constructs a typed option validation error.
func NewValidationError(field Field, rule Rule, message string) error {
	return &ValidationError{Field: field, Rule: rule, Message: message}
}

func (err *ValidationError) Error() string {
	return err.Message
}
