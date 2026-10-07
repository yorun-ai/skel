package skeleton

import (
	"strings"

	"go.yorun.ai/skel/schema"
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
	Arguments        []*schema.Argument
	Indent           int
	InputIndent      int
	ArgumentIndent   int
}

func typeView(type_ *schema.Type) *_TypeView {
	if type_ == nil {
		return nil
	}

	view := &_TypeView{Kind: "named", Nullable: type_.Nullable}
	switch type_.Kind {
	case schema.TypeKindScalar:
		view.Name = scalarName(type_.Scalar)
	case schema.TypeKindEnum:
		view.Name = type_.Enum.Name
		view.Qualifier = type_.ExternalAlias
	case schema.TypeKindData:
		view.Name = type_.Data.Name
		view.Qualifier = type_.ExternalAlias
		view.Arguments = make([]*_TypeView, 0, len(type_.TypeArguments))
		for _, argument := range type_.TypeArguments {
			view.Arguments = append(view.Arguments, typeView(argument))
		}
	case schema.TypeKindTypeParameter:
		view.Name = type_.TypeParameter.Name
	case schema.TypeKindList:
		view.Kind = "list"
		view.Value = typeView(type_.List.Element)
	case schema.TypeKindMap:
		view.Kind = "map"
		view.Key = typeView(type_.Map.Key)
		view.Value = typeView(type_.Map.Value)
	default:
		view.Name = type_.Name()
	}
	return view
}

func renderResourceCheckArguments(check *schema.ResourceCheck) []*schema.Argument {
	arguments := make([]*schema.Argument, 0, len(check.Method.Arguments))
	for _, argument := range check.Method.Arguments {
		if argument.Source == schema.ArgumentSourcePermissionCode {
			continue
		}
		arguments = append(arguments, argument)
	}
	return arguments
}

func resourceCheckView(check *schema.ResourceCheck, indent int) *_ResourceCheckView {
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

func scalarName(scalar schema.Scalar) string {
	switch scalar {
	case schema.ScalarInt:
		return "int"
	case schema.ScalarFloat:
		return "float"
	case schema.ScalarBoolean:
		return "bool"
	case schema.ScalarString:
		return "string"
	case schema.ScalarDecimal:
		return "decimal"
	case schema.ScalarBinary:
		return "binary"
	case schema.ScalarTimestamp:
		return "timestamp"
	case schema.ScalarDuration:
		return "duration"
	case schema.ScalarLocalDate:
		return "localdate"
	case schema.ScalarLocalTime:
		return "localtime"
	case schema.ScalarLocalDateTime:
		return "localdatetime"
	case schema.ScalarUUID:
		return "uuid"
	case schema.ScalarJSON:
		return "json"
	default:
		return strings.ToLower(scalar.Name())
	}
}
