package skeleton

import (
	"strings"

	"go.yorun.ai/skel/internal/model"
)

type _TypeView struct {
	Kind      string
	Name      string
	Qualifier string
	Arguments []*_TypeView
	Key       *_TypeView
	Value     *_TypeView
	Nullable  bool
}

type _ResourceCheckView struct {
	Name             string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	InputDescription string
	InputSensitive   bool
	Arguments        []*model.Argument
	Indent           int
	InputIndent      int
	ArgumentIndent   int
}

func typeView(type_ *model.Type) *_TypeView {
	if type_ == nil {
		return nil
	}

	view := &_TypeView{Kind: "named", Nullable: type_.Nullable}
	switch type_.Kind {
	case model.TypeKindScalar:
		view.Name = scalarName(type_.Scalar)
	case model.TypeKindEnum:
		view.Name = type_.Enum.Name
		view.Qualifier = type_.ExternalAlias
	case model.TypeKindData:
		view.Name = type_.Data.Name
		view.Qualifier = type_.ExternalAlias
		view.Arguments = make([]*_TypeView, 0, len(type_.TypeArguments))
		for _, argument := range type_.TypeArguments {
			view.Arguments = append(view.Arguments, typeView(argument))
		}
	case model.TypeKindTypeParameter:
		view.Name = type_.TypeParameter.Name
	case model.TypeKindList:
		view.Kind = "list"
		view.Value = typeView(type_.List.Value)
	case model.TypeKindMap:
		view.Kind = "map"
		view.Key = typeView(type_.Map.Key)
		view.Value = typeView(type_.Map.Value)
	default:
		view.Name = type_.Name()
	}
	return view
}

func renderResourceCheckArguments(check *model.ResourceCheck) []*model.Argument {
	arguments := make([]*model.Argument, 0, len(check.Method.Arguments))
	for _, argument := range check.Method.Arguments {
		if argument.Source == model.ArgumentSourcePermissionCode {
			continue
		}
		arguments = append(arguments, argument)
	}
	return arguments
}

func resourceCheckView(check *model.ResourceCheck, indent int) *_ResourceCheckView {
	return &_ResourceCheckView{
		Name:             check.Name,
		Description:      check.Method.Description,
		Deprecated:       check.Deprecated,
		DeprecatedReason: check.DeprecatedReason,
		InputDescription: check.Method.InputDescription,
		InputSensitive:   check.Method.ArgumentsSensitive,
		Arguments:        renderResourceCheckArguments(check),
		Indent:           indent,
		InputIndent:      indent + 4,
		ArgumentIndent:   indent + 8,
	}
}

func scalarName(scalar model.Scalar) string {
	switch scalar {
	case model.ScalarInt:
		return "int"
	case model.ScalarFloat:
		return "float"
	case model.ScalarBoolean:
		return "bool"
	case model.ScalarString:
		return "string"
	case model.ScalarDecimal:
		return "decimal"
	case model.ScalarBinary:
		return "binary"
	case model.ScalarTimestamp:
		return "timestamp"
	case model.ScalarDuration:
		return "duration"
	case model.ScalarLocalDate:
		return "localdate"
	case model.ScalarLocalTime:
		return "localtime"
	case model.ScalarLocalDateTime:
		return "localdatetime"
	case model.ScalarUUID:
		return "uuid"
	case model.ScalarJSON:
		return "json"
	default:
		return strings.ToLower(scalar.Name())
	}
}
