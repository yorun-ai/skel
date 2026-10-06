// Package descriptor defines runtime descriptions of Skel contracts.
//
// Descriptors contain named type references and generated contract metadata,
// without source locations, analysis state, registration, or runtime handlers.
// Language schemas are defined by go.yorun.ai/skel/schema. Descriptor producers
// project those schemas; consumers own registration and version policies.
// Methods retain declared policies alongside EffectiveAuthMode/EffectiveRequire.
// ValidateEffectivePolicy verifies only these derived values against declarations,
// using the same inheritance and composition rules as semantic schemas.
//
// Permission metadata uses full names such as PermissionRequire,
// PermissionExpression and PermissionCheckInvocation. The package does not depend
// on Vine or a compiler; bindings adapt descriptors to their runtime interfaces.
package descriptor
