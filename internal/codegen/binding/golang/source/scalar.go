package source

import (
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/schema"
)

func castScalarType(p *schema.Type) *Type {
	switch p.Scalar {
	case schema.ScalarInt:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*int", "int"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "0"),
		}
	case schema.ScalarFloat:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*float64", "float64"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "0.0"),
		}
	case schema.ScalarBoolean:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*bool", "bool"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "false"),
		}
	case schema.ScalarString:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*string", "string"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", `""`),
		}
	case schema.ScalarDecimal:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.Decimal", "types.Decimal"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.Decimal{}"),
		}
	case schema.ScalarBinary:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.Binary", "types.Binary"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: "nil",
		}
	case schema.ScalarTimestamp:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.Timestamp", "types.Timestamp"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.Timestamp{}"),
		}
	case schema.ScalarDuration:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.Duration", "types.Duration"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.Duration{}"),
		}
	case schema.ScalarLocalDate:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.LocalDate", "types.LocalDate"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.LocalDate{}"),
		}
	case schema.ScalarLocalTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.LocalTime", "types.LocalTime"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.LocalTime{}"),
		}
	case schema.ScalarLocalDateTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.LocalDateTime", "types.LocalDateTime"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.LocalDateTime{}"),
		}
	case schema.ScalarUUID:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.UUID", "types.UUID"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "types.UUID{}"),
		}
	case schema.ScalarJSON:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*types.JSON", "types.JSON"),
			Imports:      []*Import{{Path: typesImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", `""`),
		}
	}
	return nil
}
