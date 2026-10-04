package schema

import (
	"fmt"

	"go.yorun.ai/skelc/internal/util/webpath"
)

// Validate checks a schema snapshot's format version and normalized structure.
func Validate(document *Document) error {
	if document == nil {
		return fmt.Errorf("schema is required")
	}
	if document.Format != Format {
		return fmt.Errorf("unsupported schema format %q", document.Format)
	}
	if document.FormatVersion != FormatVersion {
		return fmt.Errorf("unsupported schema format version %d", document.FormatVersion)
	}
	if document.Domain == "" {
		return fmt.Errorf("schema domain is required")
	}
	if document.Declarations == nil {
		return fmt.Errorf("schema declarations are required")
	}
	seen := map[string]bool{}
	for index, declaration := range document.Declarations {
		if declaration == nil {
			return fmt.Errorf("schema declaration %d is null", index)
		}
		if declaration.Name == "" || declaration.SkelName == "" || declaration.Kind == "" {
			return fmt.Errorf("schema declaration %d has an empty name or type", index)
		}
		key := string(declaration.Kind) + "\x00" + declaration.SkelName
		if seen[key] {
			return fmt.Errorf("schema contains duplicated %s declaration %s", declaration.Kind, declaration.SkelName)
		}
		seen[key] = true
		if err := validateDeclarationBody(declaration); err != nil {
			return fmt.Errorf("schema declaration %s: %w", declaration.SkelName, err)
		}
	}
	return nil
}

func validateDeclarationBody(declaration *Declaration) error {
	present := 0
	for _, bodyPresent := range []bool{
		declaration.Enum != nil,
		declaration.Data != nil,
		declaration.Actor != nil,
		declaration.Resource != nil,
		declaration.Service != nil,
		declaration.Web != nil,
		declaration.Task != nil,
	} {
		if bodyPresent {
			present++
		}
	}
	if present != 1 {
		return fmt.Errorf("expected exactly one declaration body")
	}
	valid := false
	switch declaration.Kind {
	case DeclarationTypeEnum:
		valid = declaration.Enum != nil
	case DeclarationTypeData, DeclarationTypeConfig, DeclarationTypeEvent:
		valid = declaration.Data != nil
	case DeclarationTypeActor:
		valid = declaration.Actor != nil
	case DeclarationTypeResource:
		valid = declaration.Resource != nil
	case DeclarationTypeService:
		valid = declaration.Service != nil
		if valid && declaration.Service.Ext && (!declaration.Pub || declaration.Service.Api) {
			return fmt.Errorf("ext services must be public and cannot be api")
		}
		if valid && declaration.Pub && declaration.Service.Api {
			return fmt.Errorf("api and pub are mutually exclusive")
		}
	case DeclarationTypeWeb:
		valid = declaration.Web != nil
	case DeclarationTypeTask:
		valid = declaration.Task != nil
	}
	if !valid {
		return fmt.Errorf("declaration body does not match type %q", declaration.Kind)
	}
	switch declaration.Kind {
	case DeclarationTypeEnum:
		return validateEnum(declaration.Enum)
	case DeclarationTypeData, DeclarationTypeConfig, DeclarationTypeEvent:
		return validateDataDeclaration(declaration.Kind, declaration.Data)
	case DeclarationTypeActor:
		return validateActor(declaration.Actor)
	case DeclarationTypeResource:
		return validateResource(declaration.Resource)
	case DeclarationTypeService:
		return validateService(declaration.Service)
	case DeclarationTypeWeb:
		if declaration.Web.Audiences == nil {
			return fmt.Errorf("web audiences are required")
		}
		if declaration.Web.MountPath != "" {
			if err := webpath.Validate(declaration.Web.MountPath); err != nil {
				return fmt.Errorf("invalid web mount path: %w", err)
			}
		}
		return validateAudiences(declaration.Web.Audiences)
	case DeclarationTypeTask:
		return validateTask(declaration.Task)
	default:
		return fmt.Errorf("unsupported declaration type %q", declaration.Kind)
	}
}

func validateEnum(value *EnumSchema) error {
	if value.Items == nil {
		return fmt.Errorf("enum items are required")
	}
	seen := map[string]bool{}
	for index, item := range value.Items {
		if item == nil || item.Name == "" {
			return fmt.Errorf("enum item %d is null or unnamed", index)
		}
		if seen[item.Name] {
			return fmt.Errorf("duplicated enum item %s", item.Name)
		}
		seen[item.Name] = true
	}
	return nil
}

func validateData(value *DataSchema) error {
	return validateDataMembers(value)
}

func validateDataMembers(value *DataSchema) error {
	if value.Members == nil {
		return fmt.Errorf("data members are required")
	}
	return validateMembers(value.Members)
}

func validateDataDeclaration(kind DeclarationType, value *DataSchema) error {
	switch kind {
	case DeclarationTypeConfig:
		if value.Lifecycle != ConfigLifecycleEternal && value.Lifecycle != ConfigLifecycleInstant {
			return fmt.Errorf("unsupported config lifecycle %q", value.Lifecycle)
		}
	case DeclarationTypeData, DeclarationTypeEvent:
		if value.Lifecycle != "" {
			return fmt.Errorf("%s declaration contains config lifecycle %q", kind, value.Lifecycle)
		}
	}
	return validateDataMembers(value)
}

func validateMembers(values []*Member) error {
	seen := map[string]bool{}
	for index, member := range values {
		if member == nil || member.Name == "" {
			return fmt.Errorf("member %d is null or unnamed", index)
		}
		if seen[member.Name] {
			return fmt.Errorf("duplicated member %s", member.Name)
		}
		seen[member.Name] = true
		if err := validateType(member.Type); err != nil {
			return fmt.Errorf("member %s: %w", member.Name, err)
		}
	}
	return nil
}

func validateActor(value *ActorSchema) error {
	if value.IdentifierField != "" {
		found := false
		if value.AuthEnabled && value.AuthInfo != nil {
			for _, member := range value.AuthInfo.Members {
				if member != nil && member.Name == value.IdentifierField && member.Type != nil {
					kind := member.Type
					found = kind.Kind == TypeKindScalar && !kind.Nullable && (kind.Name == "string" || kind.Name == "uuid" || kind.Name == "int")
				}
			}
		}
		if !found {
			return fmt.Errorf("actor identifier must name a non-nullable string, uuid, or int info field")
		}
	}
	if value.Vias == nil {
		return fmt.Errorf("actor vias are required")
	}
	seen := map[string]bool{}
	for index, via := range value.Vias {
		if via == nil || via.Name == "" {
			return fmt.Errorf("actor via %d is null or unnamed", index)
		}
		if seen[via.Name] {
			return fmt.Errorf("duplicated actor via %s", via.Name)
		}
		seen[via.Name] = true
	}
	if value.AuthEnabled {
		if value.AuthCredential == nil || value.AuthInfo == nil {
			return fmt.Errorf("authenticated actor requires credential and info schemas")
		}
		if err := validateData(value.AuthCredential); err != nil {
			return fmt.Errorf("actor credential: %w", err)
		}
		if err := validateData(value.AuthInfo); err != nil {
			return fmt.Errorf("actor info: %w", err)
		}
	} else if value.AuthCredential != nil || value.AuthInfo != nil {
		return fmt.Errorf("actor without authentication contains credential or info schema")
	}
	return nil
}

func validateResource(value *ResourceSchema) error {
	if value.Actions == nil {
		return fmt.Errorf("resource actions are required")
	}
	if err := validateResourceChecks(value.Checks); err != nil {
		return err
	}
	seen := map[string]bool{}
	for index, action := range value.Actions {
		if action == nil || action.Name == "" {
			return fmt.Errorf("resource action %d is null or unnamed", index)
		}
		if seen[action.Name] {
			return fmt.Errorf("duplicated resource action %s", action.Name)
		}
		if action.PermissionCode == "" {
			return fmt.Errorf("resource action %s has no permission code", action.Name)
		}
		seen[action.Name] = true
		if err := validateResourceChecks(action.Checks); err != nil {
			return fmt.Errorf("resource action %s: %w", action.Name, err)
		}
	}
	return nil
}

func validateResourceChecks(values []*ResourceCheck) error {
	seen := map[string]bool{}
	for index, check := range values {
		if check == nil || check.Name == "" {
			return fmt.Errorf("resource check %d is null or unnamed", index)
		}
		if seen[check.Name] {
			return fmt.Errorf("duplicated resource check %s", check.Name)
		}
		if check.Arguments == nil {
			return fmt.Errorf("resource check %s arguments are required", check.Name)
		}
		seen[check.Name] = true
		if err := validateArguments(check.Arguments); err != nil {
			return fmt.Errorf("resource check %s: %w", check.Name, err)
		}
	}
	return nil
}
