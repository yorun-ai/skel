package schema

import "fmt"

func validateRequirement(value *Requirement) error {
	if value == nil {
		return nil
	}
	return validateRequirementPath(value, map[*Requirement]uint8{})
}

func validateRequirementPath(value *Requirement, active map[*Requirement]uint8) error {
	if value == nil {
		return fmt.Errorf("permission requirement child is null")
	}
	if active[value] == 1 {
		return fmt.Errorf("cyclic permission requirement")
	}
	if active[value] == 2 {
		return nil
	}
	active[value] = 1
	switch value.Mode {
	case RequirementModeCode:
		if value.Code == "" {
			return fmt.Errorf("code permission requirement has no code")
		}
		if value.Check != nil || len(value.Children) != 0 {
			return fmt.Errorf("code permission requirement contains unrelated fields")
		}
	case RequirementModeReference:
		if value.Check == nil || value.Check.Resource == "" || value.Check.Action == "" {
			return fmt.Errorf("reference permission requirement is incomplete")
		}
		if value.Code != "" || len(value.Children) != 0 {
			return fmt.Errorf("reference permission requirement contains unrelated fields")
		}
		for _, argument := range value.Check.Arguments {
			if argument == nil || argument.Name == "" && argument.JSONPath == "" {
				return fmt.Errorf("reference permission requirement has an unnamed argument")
			}
			if argument.Type != nil {
				if err := validateType(argument.Type); err != nil {
					return err
				}
			}
		}
	case RequirementModeCheck:
		if value.Check == nil || value.Check.Resource == "" || value.Check.Check == "" {
			return fmt.Errorf("check permission requirement is incomplete")
		}
		if value.Code != "" || len(value.Children) != 0 {
			return fmt.Errorf("check permission requirement contains unrelated fields")
		}
		for _, argument := range value.Check.Arguments {
			if argument == nil || argument.Name == "" {
				return fmt.Errorf("check permission requirement has an unnamed argument")
			}
			if err := validateType(argument.Type); err != nil {
				return err
			}
		}
	case RequirementModeAll, RequirementModeAny:
		if len(value.Children) == 0 {
			return fmt.Errorf("%s permission requirement has no children", value.Mode)
		}
		if value.Code != "" || value.Check != nil {
			return fmt.Errorf("%s permission requirement contains unrelated fields", value.Mode)
		}
		for _, child := range value.Children {
			if err := validateRequirementPath(child, active); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported permission requirement mode %q", value.Mode)
	}
	active[value] = 2
	return nil
}

func validateTypeOptional(value *Type) error {
	if value == nil {
		return nil
	}
	return validateType(value)
}

func validateType(value *Type) error {
	return validateTypePath(value, map[*Type]uint8{})
}

func validateTypePath(value *Type, active map[*Type]uint8) error {
	if value == nil {
		return fmt.Errorf("type is required")
	}
	if active[value] == 1 {
		return fmt.Errorf("cyclic type structure")
	}
	if active[value] == 2 {
		return nil
	}
	active[value] = 1
	switch value.Kind {
	case TypeKindScalar, TypeKindEnum, TypeKindData, TypeKindConfig, TypeKindEvent, TypeKindTypeParameter, TypeKindImportedReference:
		if value.Name == "" {
			return fmt.Errorf("%s type has no name", value.Kind)
		}
		if value.Element != nil || value.Key != nil || value.Value != nil {
			return fmt.Errorf("%s type contains unrelated fields", value.Kind)
		}
		if value.Kind != TypeKindData && value.Kind != TypeKindImportedReference && len(value.Arguments) != 0 {
			return fmt.Errorf("%s type contains type arguments", value.Kind)
		}
	case TypeKindList:
		if value.Name != "" || len(value.Arguments) != 0 || value.Key != nil || value.Value != nil {
			return fmt.Errorf("list type contains unrelated fields")
		}
		if err := validateTypePath(value.Element, active); err != nil {
			return fmt.Errorf("list element: %w", err)
		}
	case TypeKindMap:
		if value.Name != "" || len(value.Arguments) != 0 || value.Element != nil {
			return fmt.Errorf("map type contains unrelated fields")
		}
		if err := validateTypePath(value.Key, active); err != nil {
			return fmt.Errorf("map key: %w", err)
		}
		if err := validateTypePath(value.Value, active); err != nil {
			return fmt.Errorf("map value: %w", err)
		}
	default:
		return fmt.Errorf("unsupported type kind %q", value.Kind)
	}
	for index, argument := range value.Arguments {
		if err := validateTypePath(argument, active); err != nil {
			return fmt.Errorf("type argument %d: %w", index, err)
		}
	}
	active[value] = 2
	return nil
}
