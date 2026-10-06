package vineschema

import (
	"strings"

	"go.yorun.ai/skel/internal/model"
	contractschema "go.yorun.ai/skel/schema"
)

func (g *_Gen) buildEnumSchema(value *model.Enum, projected *contractschema.Declaration) *_EnumSchema {
	result := &_EnumSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Items: make([]*_EnumItemSchema, 0, len(projected.Enum.Items)),
	}
	for _, item := range projected.Enum.Items {
		result.Items = append(result.Items, &_EnumItemSchema{
			Name: item.Name, Description: item.Description,
			Deprecated: item.Deprecated, DeprecatedReason: item.DeprecatedReason,
		})
	}
	return result
}

func (g *_Gen) buildDataSchema(value *model.Data, projected *contractschema.Declaration) *_DataSchema {
	return &_DataSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Sensitive: projected.Data.Sensitive,
		TypeParameters: append([]string{}, projected.Data.TypeParameters...),
		Members:        g.buildMemberSchemas(projected.Data.Members),
	}
}

func (g *_Gen) buildConfigSchema(value *model.Data, projected *contractschema.Declaration) *_ConfigSchema {
	return &_ConfigSchema{
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Pub: projected.Pub,
		Sensitive: projected.Data.Sensitive, Lifecycle: strings.ToUpper(string(projected.Data.Lifecycle)),
		Members: g.buildMemberSchemas(projected.Data.Members),
	}
}

func (g *_Gen) buildEventSchema(value *model.Data, projected *contractschema.Declaration) *_EventSchema {
	return &_EventSchema{
		Ext:  projected.Data.Ext,
		Name: projected.Name, SkelName: projected.SkelName, Hash: value.Hash,
		Description: projected.Description, Deprecated: projected.Deprecated,
		DeprecatedReason: projected.DeprecatedReason, Pub: projected.Pub,
		Sensitive: projected.Data.Sensitive, Members: g.buildMemberSchemas(projected.Data.Members),
	}
}
