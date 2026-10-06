package descriptor

import (
	"strings"

	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildMemberDescriptors(members []*schema.DataMember) []*descriptor.Member {
	schemas := make([]*descriptor.Member, 0, len(members))
	for _, member := range members {
		schemas = append(schemas, &descriptor.Member{
			Name: member.Name, Description: member.Description,
			Deprecated: member.Deprecated, DeprecatedReason: member.DeprecatedReason,
			Example: member.Example, Sensitive: member.Sensitive, Type: g.buildTypeDescriptor(member.Type),
		})
	}
	return schemas
}

func (g *_Gen) buildArgumentDescriptors(arguments []*schema.Argument) []*descriptor.Member {
	schemas := make([]*descriptor.Member, 0, len(arguments))
	for _, argument := range arguments {
		schemas = append(schemas, &descriptor.Member{
			Name: argument.Name, Description: argument.Description,
			Deprecated: argument.Deprecated, DeprecatedReason: argument.DeprecatedReason,
			Example: argument.Example, Sensitive: argument.Sensitive, Type: g.buildTypeDescriptor(argument.Type),
		})
	}
	return schemas
}

func (g *_Gen) buildTypeDescriptor(value *schema.Type) *descriptor.Type {
	if value == nil {
		return nil
	}
	result := &descriptor.Type{Nullable: value.Nullable}
	switch value.Kind {
	case schema.TypeKindScalar:
		result.Kind = descriptor.TypeKindScalar
		result.Scalar = descriptor.Scalar(strings.ToLower(value.Scalar.Name()))
	case schema.TypeKindList:
		result.Kind = descriptor.TypeKindList
		result.Element = g.buildTypeDescriptor(value.List.Value)
	case schema.TypeKindMap:
		result.Kind = descriptor.TypeKindMap
		result.Key = g.buildTypeDescriptor(value.Map.Key)
		result.Value = g.buildTypeDescriptor(value.Map.Value)
	case schema.TypeKindEnum:
		result.Kind = descriptor.TypeKindEnum
		result.Name, result.SkelName = localAndSkelName(g.Domain.TypeReferenceName(value))
	case schema.TypeKindData:
		result.Kind = descriptor.TypeKind(value.Data.Kind)
		result.Name, result.SkelName = localAndSkelName(g.Domain.TypeReferenceName(value))
	case schema.TypeKindTypeParameter:
		result.Kind = descriptor.TypeKindTypeParameter
		result.Name = value.TypeParameter.Name
	default:
		return nil
	}
	for _, argument := range value.TypeArguments {
		result.TypeArguments = append(result.TypeArguments, g.buildTypeDescriptor(argument))
	}
	return result
}

func localAndSkelName(skelName string) (string, string) {
	index := strings.LastIndex(skelName, ".")
	if index < 0 {
		return skelName, skelName
	}
	return skelName[index+1:], skelName
}

func typeParameterNames(values []*schema.TypeParameter) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}
