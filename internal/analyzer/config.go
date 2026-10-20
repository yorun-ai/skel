package analyzer

import "go.yorun.ai/skel/schema"

// Resolve config adoption after field types so forward references and imported
// members preserve their defining domain rather than being resolved again.
func (p *Analysis) normalizeConfigSources(allowMissing bool) {
	for _, config := range sortData(p.dataMap) {
		ref := config.From
		if ref == nil {
			continue
		}
		source := p.dataMap[ref.SkelName]
		if ref.ExternalAlias != "" {
			imported := p.importsMap[ref.ExternalAlias]
			if imported == nil {
				declared := false
				for _, item := range p.content.Imports {
					qualifier := item.Domain.String()
					if item.Alias != nil {
						qualifier = item.Alias.Value
					}
					if qualifier == ref.ExternalAlias {
						declared = true
						ref.ExternalDomain = item.Domain.String()
					}
				}
				if allowMissing && declared {
					continue
				}
				p.reporter.reportf("%s import %s not found", ref.Pos, ref.ExternalAlias)
				continue
			}
			source = imported.Domain.dataMap[ref.SkelName]
			ref.ExternalDomain = imported.Domain.name
			ref.ExternalAliasExplicit = imported.Schema.ExplicitAlias
		}
		if !p.reporter.checkReference(source != nil, "%s config source %s not found", ref.Pos, ref.SkelName) {
			continue
		}
		if !p.reporter.check(source.Kind == schema.DataKindConfig && source.Abs && source.From == nil,
			"%s config source %s must be an abstract config defined with a body", ref.Pos, source.SkelName) {
			continue
		}
		if p.invalidData[source] {
			p.invalidData[config] = true
			continue
		}
		ref.Kind, ref.Data, ref.SkelName = schema.TypeKindData, source, source.SkelName
		config.Lifecycle = source.Lifecycle
		config.Sensitive = config.Sensitive || source.Sensitive
		config.Members = make([]*schema.DataMember, len(source.Members))
		for i, member := range source.Members {
			copy := *member
			copy.Type = p.configMemberType(member.Type)
			config.Members[i] = &copy
		}
	}
}

func (p *Analysis) configMemberType(kind *schema.Type) *schema.Type {
	if kind == nil {
		return nil
	}
	copy := *kind
	domain := ""
	if kind.Data != nil {
		domain = kind.Data.Domain
	}
	if kind.Enum != nil {
		domain = kind.Enum.Domain
	}
	if domain != "" {
		copy.ExternalDomain, copy.ExternalAlias, copy.ExternalAliasExplicit = "", "", false
		if domain != p.name {
			copy.ExternalDomain, copy.ExternalAlias = domain, domain
			for _, imported := range p.imports {
				if imported.Name == domain {
					copy.ExternalAlias, copy.ExternalAliasExplicit = imported.Alias, imported.ExplicitAlias
					break
				}
			}
		}
	}
	if kind.List != nil {
		copy.List = &schema.ListType{Element: p.configMemberType(kind.List.Element)}
	}
	if kind.Map != nil {
		copy.Map = &schema.MapType{Key: p.configMemberType(kind.Map.Key), Value: p.configMemberType(kind.Map.Value)}
	}
	if kind.TypeArguments != nil {
		copy.TypeArguments = make([]*schema.Type, len(kind.TypeArguments))
		for i, arg := range kind.TypeArguments {
			copy.TypeArguments[i] = p.configMemberType(arg)
		}
	}
	return &copy
}
