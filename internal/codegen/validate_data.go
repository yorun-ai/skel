package codegen

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

func validateData(data *model.Data) error {
	return validateDataGraph(data, map[*model.Data]bool{})
}

func validateDataGraph(data *model.Data, seen map[*model.Data]bool) error {
	if data == nil {
		return fmt.Errorf("generated model contains nil data")
	}
	if data.Ext && (data.Kind != model.DataKindEvent || data.Pub) {
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
		if err := validateModelTypeGraph(member.Type, seen); err != nil {
			return fmt.Errorf("data %s member %s: %w", data.Name, member.Name, err)
		}
	}
	return nil
}

func validateModelType(type_ *model.Type) error {
	return validateModelTypeGraph(type_, map[*model.Data]bool{})
}

func validateModelTypeGraph(type_ *model.Type, seenData map[*model.Data]bool) error {
	if type_ == nil {
		return fmt.Errorf("type is nil")
	}
	active := map[*model.Type]bool{}
	checked := map[*model.Type]bool{}
	var validate func(*model.Type) error
	validate = func(current *model.Type) error {
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
		if current.Kind != model.TypeKindData && len(current.TypeArguments) != 0 {
			return fmt.Errorf("type kind %d does not support type arguments", current.Kind)
		}
		switch current.Kind {
		case model.TypeKindScalar:
			if current.Scalar < model.ScalarInt || current.Scalar > model.ScalarJSON {
				return fmt.Errorf("unsupported scalar %s", current.Scalar.Name())
			}
		case model.TypeKindList:
			if current.List == nil || current.List.Value == nil {
				return fmt.Errorf("list metadata is nil")
			}
		case model.TypeKindMap:
			if current.Map == nil || current.Map.Key == nil || current.Map.Value == nil {
				return fmt.Errorf("map metadata is nil")
			}
		case model.TypeKindEnum:
			if current.Enum == nil {
				return fmt.Errorf("enum metadata is nil")
			}
		case model.TypeKindData:
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
			case model.DataKindData, model.DataKindConfig, model.DataKindEvent:
			default:
				return fmt.Errorf("referenced data %s has unsupported kind %q", current.Data.Name, current.Data.Kind)
			}
			if err := validateDataGraph(current.Data, seenData); err != nil {
				return err
			}
		case model.TypeKindTypeParameter:
			if current.TypeParameter == nil {
				return fmt.Errorf("type parameter metadata is nil")
			}
		default:
			return fmt.Errorf("unsupported type kind %d", current.Kind)
		}
		children := current.TypeArguments
		switch current.Kind {
		case model.TypeKindList:
			children = []*model.Type{current.List.Value}
		case model.TypeKindMap:
			children = []*model.Type{current.Map.Key, current.Map.Value}
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
