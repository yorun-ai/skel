package descriptor

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/skel/schema"
)

func (g *_Gen) buildEnumDescriptor(value *schema.Enum) *descriptor.Enum {
	result := &descriptor.Enum{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Items: make([]*descriptor.EnumItem, 0, len(value.Items)),
	}
	for _, item := range value.Items {
		result.Items = append(result.Items, &descriptor.EnumItem{
			Name: item.Name, Description: item.Description,
			Deprecated: item.Deprecated, DeprecatedReason: item.DeprecatedReason,
		})
	}
	return result
}

func (g *_Gen) buildDataDescriptor(value *schema.Data) *descriptor.Data {
	return &descriptor.Data{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Sensitive: value.Sensitive,
		TypeParameters: typeParameterNames(value.TypeParameters),
		Members:        g.buildMemberDescriptors(value.Members),
	}
}

func (g *_Gen) buildConfigDescriptor(value *schema.Data) *descriptor.Config {
	return &descriptor.Config{
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Pub: value.Pub,
		Sensitive: value.Sensitive, Lifecycle: descriptor.ConfigLifecycle(value.Lifecycle),
		Members: g.buildMemberDescriptors(value.Members),
	}
}

func (g *_Gen) buildEventDescriptor(value *schema.Data) *descriptor.Event {
	return &descriptor.Event{
		Ext:  value.Ext,
		Name: value.Name, SkelName: value.SkelName, Hash: value.Hash,
		Description: value.Description, Deprecated: value.Deprecated,
		DeprecatedReason: value.DeprecatedReason, Pub: value.Pub,
		Sensitive: value.Sensitive, Members: g.buildMemberDescriptors(value.Members),
	}
}
