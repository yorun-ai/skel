package source

import (
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/model"
)

func castScalarType(p *model.Type) *Type {
	switch p.Scalar {
	case model.ScalarInt:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*int", "int"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "0"),
		}
	case model.ScalarFloat:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*float64", "float64"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "0.0"),
		}
	case model.ScalarBoolean:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*bool", "bool"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "false"),
		}
	case model.ScalarString:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*string", "string"),
			DefaultValue: binding.ChooseString(p.Nullable, "nil", `""`),
		}
	case model.ScalarDecimal:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.Decimal", "skel.Decimal"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.Decimal{}"),
		}
	case model.ScalarBinary:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.Binary", "skel.Binary"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: "nil",
		}
	case model.ScalarTimestamp:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.Timestamp", "skel.Timestamp"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.Timestamp{}"),
		}
	case model.ScalarDuration:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.Duration", "skel.Duration"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.Duration{}"),
		}
	case model.ScalarLocalDate:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.LocalDate", "skel.LocalDate"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.LocalDate{}"),
		}
	case model.ScalarLocalTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.LocalTime", "skel.LocalTime"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.LocalTime{}"),
		}
	case model.ScalarLocalDateTime:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.LocalDateTime", "skel.LocalDateTime"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.LocalDateTime{}"),
		}
	case model.ScalarUUID:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.UUID", "skel.UUID"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", "skel.UUID{}"),
		}
	case model.ScalarJSON:
		return &Type{
			Plain:        binding.ChooseString(p.Nullable, "*skel.JSON", "skel.JSON"),
			Imports:      []*Import{{Path: skelImport}},
			DefaultValue: binding.ChooseString(p.Nullable, "nil", `""`),
		}
	}
	return nil
}
