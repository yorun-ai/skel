package vineschema

import (
	"strings"

	contractschema "go.yorun.ai/skel/internal/schema"
)

func (g *_Gen) buildMemberSchemas(members []*contractschema.Member) []*_MemberSchema {
	schemas := make([]*_MemberSchema, 0, len(members))
	for _, member := range members {
		schemas = append(schemas, &_MemberSchema{
			Name: member.Name, Description: member.Description,
			Deprecated: member.Deprecated, DeprecatedReason: member.DeprecatedReason,
			Example: member.Example, Sensitive: member.Sensitive, Type: g.buildTypeSchema(member.Type),
		})
	}
	return schemas
}

func (g *_Gen) buildArgumentSchemas(arguments []*contractschema.Argument) []*_MemberSchema {
	schemas := make([]*_MemberSchema, 0, len(arguments))
	for _, argument := range arguments {
		schemas = append(schemas, &_MemberSchema{
			Name: argument.Name, Description: argument.Description,
			Deprecated: argument.Deprecated, DeprecatedReason: argument.DeprecatedReason,
			Example: argument.Example, Sensitive: argument.Sensitive, Type: g.buildTypeSchema(argument.Type),
		})
	}
	return schemas
}

func (g *_Gen) buildTypeSchema(value *contractschema.Type) *_TypeSchema {
	if value == nil {
		return nil
	}
	result := &_TypeSchema{Nullable: value.Nullable}
	switch value.Kind {
	case contractschema.TypeKindScalar:
		result.Kind = typeKindScalar
		result.Scalar = _Scalar(value.Name)
	case contractschema.TypeKindList:
		result.Kind = typeKindList
		result.Element = g.buildTypeSchema(value.Element)
	case contractschema.TypeKindMap:
		result.Kind = typeKindMap
		result.Key = g.buildTypeSchema(value.Key)
		result.Value = g.buildTypeSchema(value.Value)
	case contractschema.TypeKindEnum:
		result.Kind = typeKindEnum
		result.Name, result.SkelName = localAndSkelName(value.Name)
	case contractschema.TypeKindData:
		result.Kind = typeKindData
		result.Name, result.SkelName = localAndSkelName(value.Name)
	case contractschema.TypeKindConfig:
		result.Kind = typeKindConfig
		result.Name, result.SkelName = localAndSkelName(value.Name)
	case contractschema.TypeKindEvent:
		result.Kind = typeKindEvent
		result.Name, result.SkelName = localAndSkelName(value.Name)
	case contractschema.TypeKindTypeParameter:
		result.Kind = typeKindTypeParameter
		result.Name = value.Name
	default:
		return nil
	}
	for _, argument := range value.Arguments {
		result.TypeArguments = append(result.TypeArguments, g.buildTypeSchema(argument))
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
