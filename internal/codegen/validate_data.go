package codegen

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

func validateData(data *schema.Data) error {
	return validateDataGraph(data, map[*schema.Data]bool{})
}

func validateDataGraph(data *schema.Data, seen map[*schema.Data]bool) error {
	if data == nil {
		return fmt.Errorf("generated schema contains nil data")
	}
	if data.Ext && (data.Kind != schema.DataKindEvent || data.Pub) {
		return fmt.Errorf("ext is only allowed on events and cannot be combined with pub")
	}
	if seen[data] {
		return nil
	}
	seen[data] = true
	for _, parameter := range data.TypeParameters {
		if parameter == nil {
			return fmt.Errorf("data %s contains a nil type parameter", data.Name)
		}
	}
	for _, member := range data.Members {
		if member == nil {
			return fmt.Errorf("data %s contains a nil member", data.Name)
		}
		if err := validateSchemaTypeGraph(member.Type, seen); err != nil {
			return fmt.Errorf("data %s member %s: %w", data.Name, member.Name, err)
		}
	}
	return nil
}

func validateSchemaType(type_ *schema.Type) error {
	return validateSchemaTypeGraph(type_, map[*schema.Data]bool{})
}

func validateSchemaTypeGraph(type_ *schema.Type, seenData map[*schema.Data]bool) error {
	if type_ == nil {
		return fmt.Errorf("type is nil")
	}
	active := map[*schema.Type]bool{}
	checked := map[*schema.Type]bool{}
	var validate func(*schema.Type) error
	validate = func(current *schema.Type) error {
		if current == nil {
			return fmt.Errorf("type is nil")
		}
		if active[current] {
			return fmt.Errorf("cyclic type structure")
		}
		if checked[current] {
			return nil
		}
		active[current] = true
		defer delete(active, current)
		if current.Kind != schema.TypeKindData && len(current.TypeArguments) != 0 {
			return fmt.Errorf("type kind %d does not support type arguments", current.Kind)
		}
		switch current.Kind {
		case schema.TypeKindScalar:
			if current.Scalar < schema.ScalarInt || current.Scalar > schema.ScalarJSON {
				return fmt.Errorf("unsupported scalar %s", current.Scalar.Name())
			}
		case schema.TypeKindList:
			if current.List == nil || current.List.Value == nil {
				return fmt.Errorf("list metadata is nil")
			}
		case schema.TypeKindMap:
			if current.Map == nil || current.Map.Key == nil || current.Map.Value == nil {
				return fmt.Errorf("map metadata is nil")
			}
		case schema.TypeKindEnum:
			if current.Enum == nil {
				return fmt.Errorf("enum metadata is nil")
			}
		case schema.TypeKindData:
			if current.Data == nil {
				return fmt.Errorf("data metadata is nil")
			}
			for _, parameter := range current.Data.TypeParameters {
				if parameter == nil {
					return fmt.Errorf("data %s contains a nil type parameter", current.Data.Name)
				}
			}
			if len(current.TypeArguments) != len(current.Data.TypeParameters) {
				return fmt.Errorf("data %s has mismatched type arguments: found=%d, expected=%d", current.Data.Name, len(current.TypeArguments), len(current.Data.TypeParameters))
			}
			switch current.Data.Kind {
			case schema.DataKindData, schema.DataKindConfig, schema.DataKindEvent:
			default:
				return fmt.Errorf("referenced data %s has unsupported kind %q", current.Data.Name, current.Data.Kind)
			}
			if err := validateDataGraph(current.Data, seenData); err != nil {
				return err
			}
		case schema.TypeKindTypeParameter:
			if current.TypeParameter == nil {
				return fmt.Errorf("type parameter metadata is nil")
			}
		default:
			return fmt.Errorf("unsupported type kind %d", current.Kind)
		}
		children := current.TypeArguments
		switch current.Kind {
		case schema.TypeKindList:
			children = []*schema.Type{current.List.Value}
		case schema.TypeKindMap:
			children = []*schema.Type{current.Map.Key, current.Map.Value}
		}
		for _, child := range children {
			if err := validate(child); err != nil {
				return err
			}
		}
		checked[current] = true
		return nil
	}
	return validate(type_)
}
