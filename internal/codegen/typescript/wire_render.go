package typescript

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

func (b *_WireSchemaBuilder) renderFactories() []*_WireFactory {
	factories := make([]*_WireFactory, 0, len(b.data))
	for _, dataType := range b.sortedData() {
		parameters := make([]string, 0, len(dataType.TypeParameters))
		for _, parameter := range dataType.TypeParameters {
			parameters = append(parameters, wireTypeParameterName(parameter)+": VrpcWireSchema")
		}

		var signature string
		if len(parameters) == 0 {
			signature = "()"
		} else {
			signature = "(\n  " + strings.Join(parameters, ",\n  ") + ",\n)"
		}

		code := fmt.Sprintf(
			"function %s%s: VrpcWireSchema {\n  return %s;\n}",
			b.factoryNames[dataType],
			signature,
			b.renderObjectSchema(dataType.Members, 1),
		)
		factories = append(factories, &_WireFactory{Code: code})
	}
	return factories
}

func (b *_WireSchemaBuilder) renderMethod(method *model.Method) *_WireMethod {
	wireMethod := &_WireMethod{Name: nameutil.ToLowerCamel(method.Name)}
	if methodArgumentsContainBinary(method) {
		members := make([]*model.DataMember, 0, len(method.Arguments))
		for _, argument := range method.Arguments {
			members = append(members, &model.DataMember{Name: nameutil.ToLowerCamel(argument.Name), Type: argument.Type})
		}
		wireMethod.ArgumentsSchema = b.renderObjectSchema(members, 3)
	}
	if methodResultContainsBinary(method) {
		wireMethod.ResultSchema = b.renderType(method.ResultType, 3)
	}
	return wireMethod
}

func (b *_WireSchemaBuilder) renderObjectSchema(members []*model.DataMember, depth int) string {
	var rendered strings.Builder
	rendered.WriteString("{\n")
	rendered.WriteString(indentWire(depth + 1))
	rendered.WriteString("kind: 'object',\n")
	rendered.WriteString(indentWire(depth + 1))
	rendered.WriteString("fields: () => ({")
	if len(members) > 0 {
		rendered.WriteString("\n")
		for _, member := range members {
			rendered.WriteString(indentWire(depth + 2))
			rendered.WriteString(nameutil.ToLowerCamel(member.Name))
			rendered.WriteString(": ")
			rendered.WriteString(b.renderType(member.Type, depth+2))
			rendered.WriteString(",\n")
		}
		rendered.WriteString(indentWire(depth + 1))
	}
	rendered.WriteString("}),\n")
	rendered.WriteString(indentWire(depth))
	rendered.WriteString("}")
	return rendered.String()
}

func (b *_WireSchemaBuilder) renderType(type_ *model.Type, depth int) string {
	if type_ == nil {
		b.fail("cannot render nil TypeScript wire schema")
		return ""
	}

	var rendered string
	switch type_.Kind {
	case model.TypeKindScalar:
		kind := "value"
		if type_.Scalar == model.ScalarBinary {
			kind = "binary"
		}
		rendered = renderSimpleWireSchema(kind, type_.Nullable)
	case model.TypeKindEnum:
		rendered = renderSimpleWireSchema("value", type_.Nullable)
	case model.TypeKindTypeParameter:
		rendered = wireTypeParameterName(type_.TypeParameter)
		if type_.Nullable {
			rendered = "{ ..." + rendered + ", nullable: true }"
		}
	case model.TypeKindList:
		rendered = b.renderContainerSchema(
			"list",
			[]string{"value: " + b.renderType(type_.List.Value, depth+1)},
			type_.Nullable,
			depth,
		)
	case model.TypeKindMap:
		key := "string"
		if type_.Map.Key.Kind == model.TypeKindScalar && type_.Map.Key.Scalar == model.ScalarInt {
			key = "int"
		}
		rendered = b.renderContainerSchema(
			"map",
			[]string{
				fmt.Sprintf("key: '%s'", key),
				"value: " + b.renderType(type_.Map.Value, depth+1),
			},
			type_.Nullable,
			depth,
		)
	case model.TypeKindData:
		arguments := make([]string, 0, len(type_.TypeArguments))
		for _, argument := range type_.TypeArguments {
			arguments = append(arguments, b.renderType(argument, depth))
		}
		name := b.factoryNames[type_.Data]
		if name == "" {
			b.fail("missing TypeScript wire schema factory for %s", type_.Data.Name)
			return ""
		}
		rendered = fmt.Sprintf("%s(%s)", name, strings.Join(arguments, ", "))
		if type_.Nullable {
			rendered = "{ ..." + rendered + ", nullable: true }"
		}
	default:
		b.fail("unsupported TypeScript wire schema type %+v", type_)
	}
	return rendered
}

func (b *_WireSchemaBuilder) renderContainerSchema(
	kind string,
	fields []string,
	nullable bool,
	depth int,
) string {
	var rendered strings.Builder
	rendered.WriteString("{\n")
	rendered.WriteString(indentWire(depth + 1))
	rendered.WriteString(fmt.Sprintf("kind: '%s',\n", kind))
	for _, field := range fields {
		rendered.WriteString(indentWire(depth + 1))
		rendered.WriteString(field)
		rendered.WriteString(",\n")
	}
	if nullable {
		rendered.WriteString(indentWire(depth + 1))
		rendered.WriteString("nullable: true,\n")
	}
	rendered.WriteString(indentWire(depth))
	rendered.WriteString("}")
	return rendered.String()
}

func renderSimpleWireSchema(kind string, nullable bool) string {
	if nullable {
		return fmt.Sprintf("{ kind: '%s', nullable: true }", kind)
	}
	return fmt.Sprintf("{ kind: '%s' }", kind)
}

func wireTypeParameterName(parameter *model.TypeParameter) string {
	return nameutil.ToLowerCamel(parameter.Name) + "WireSchema"
}

func indentWire(depth int) string {
	return strings.Repeat("  ", depth)
}
