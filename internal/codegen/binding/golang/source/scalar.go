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
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.Decimal", "skeltype.Decimal"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.Decimal{}"),
		}
	case schema.ScalarBinary:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.Binary", "skeltype.Binary"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: "nil",
		}
	case schema.ScalarTimestamp:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.Timestamp", "skeltype.Timestamp"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.Timestamp{}"),
		}
	case schema.ScalarDuration:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.Duration", "skeltype.Duration"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.Duration{}"),
		}
	case schema.ScalarLocalDate:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.LocalDate", "skeltype.LocalDate"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.LocalDate{}"),
		}
	case schema.ScalarLocalTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.LocalTime", "skeltype.LocalTime"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.LocalTime{}"),
		}
	case schema.ScalarLocalDateTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.LocalDateTime", "skeltype.LocalDateTime"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.LocalDateTime{}"),
		}
	case schema.ScalarUUID:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.UUID", "skeltype.UUID"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skeltype.UUID{}"),
		}
	case schema.ScalarJSON:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skeltype.JSON", "skeltype.JSON"),
			Imports:      []*Import{{Path: typesImport, Alias: "skeltype"}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", `""`),
		}
	}
	return nil
}
