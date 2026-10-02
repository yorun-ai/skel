package analyzer

import (
	"fmt"
	"maps"
	"slices"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/nameutil"
	"go.yorun.ai/skelc/internal/util/sliceutil"
)

func (p *Analysis) normalize() {
	p.normalizeWithMissingImports(false)
}

// normalizeImport resolves references owned by the imported domain while
// preserving qualified references whose transitive domains were not loaded.
func (p *Analysis) normalizeImport() {
	p.normalizeWithMissingImports(true)
}

func (p *Analysis) normalizeWithMissingImports(allowMissingImports bool) {
	refs := &_RefContext{
		enums:                  p.enumsMap,
		dataList:               p.dataMap,
		imports:                p.importsMap,
		invalidData:            p.invalidData,
		unavailable:            p.unavailable,
		allowUnresolvedImports: allowMissingImports,
	}
	p.normalizeDeclaredData(refs)
	p.normalizeOwnedTypes(refs, allowMissingImports)
	p.validateNormalizedData()
}

func (p *Analysis) normalizeDeclaredData(refs *_RefContext) {
	for _, name := range slices.Sorted(maps.Keys(p.dataMap)) {
		dataType := p.dataMap[name]
		if !p.normalizeDataType(dataType, refs) {
			p.invalidData[dataType] = true
			p.unavailable[dataType.Name] = true
		}
	}
	p.propagateInvalidData()
}

func (p *Analysis) normalizeOwnedTypes(refs *_RefContext, allowMissingImports bool) {
	for _, name := range slices.Sorted(maps.Keys(p.actorsMap)) {
		actor := p.actorsMap[name]
		p.normalizeDataType(actor.AuthCredential, refs)
		p.normalizeDataType(actor.AuthInfo, refs)
		if actor.PermService != nil {
			p.normalizeServiceTypes(actor.PermService, refs)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.resourcesMap)) {
		resource := p.resourcesMap[name]
		if resource.CheckService != nil {
			p.normalizeServiceTypes(resource.CheckService, refs)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.servicesMap)) {
		service := p.servicesMap[name]
		valid := p.normalizeServiceTypes(service, refs)
		if allowMissingImports {
			continue
		}
		valid = p.checkActorAudiences(service.Audiences, service.Pos, "service", service.Name) && valid
		if valid {
			p.normalizeServiceRequire(service)
		}
	}
	if !allowMissingImports {
		for _, name := range slices.Sorted(maps.Keys(p.websMap)) {
			web := p.websMap[name]
			p.checkActorAudiences(web.Audiences, web.Pos, "web", web.Name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.tasksMap)) {
		task := p.tasksMap[name]
		for _, trigger := range task.Triggers {
			for _, arg := range trigger.Arguments {
				fixTypeRef(p.reporter, arg.Type, refs)
			}
		}
	}
}

func (p *Analysis) validateNormalizedData() {
	allData := sliceutil.Filter(sortData(p.dataMap), func(dataType *model.Data) bool {
		return !p.invalidData[dataType]
	})
	for _, name := range slices.Sorted(maps.Keys(p.actorsMap)) {
		actor := p.actorsMap[name]
		if actor.AuthCredential != nil {
			allData = append(allData, actor.AuthCredential)
		}
		if actor.AuthInfo != nil {
			allData = append(allData, actor.AuthInfo)
		}
	}
	p.checkHardCycleReferences(allData)
}

func (p *Analysis) normalizeDataType(dataType *model.Data, refs *_RefContext) bool {
	if dataType == nil {
		return true
	}
	refs.typeParameters = sliceutil.MapToMap(dataType.TypeParameters, func(typeParam *model.TypeParameter) (string, *model.TypeParameter) {
		return typeParam.Name, typeParam
	})
	defer func() {
		refs.typeParameters = nil
	}()

	valid := true
	for _, member := range dataType.Members {
		valid = fixTypeRef(p.reporter, member.Type, refs) && valid
	}
	return valid
}

func (p *Analysis) propagateInvalidData() {
	for changed := true; changed; {
		changed = false
		for _, dataType := range p.dataMap {
			if p.invalidData[dataType] {
				continue
			}
			for _, member := range dataType.Members {
				if referencesInvalidData(member.Type, p.invalidData) {
					p.invalidData[dataType] = true
					p.unavailable[dataType.Name] = true
					changed = true
					break
				}
			}
		}
	}
}

func referencesInvalidData(type_ *model.Type, invalid map[*model.Data]bool) bool {
	if type_ == nil {
		return false
	}
	switch type_.Kind {
	case model.TypeKindData:
		if invalid[type_.Data] {
			return true
		}
		for _, argument := range type_.TypeArguments {
			if referencesInvalidData(argument, invalid) {
				return true
			}
		}
	case model.TypeKindList:
		return referencesInvalidData(type_.List.Value, invalid)
	case model.TypeKindMap:
		return referencesInvalidData(type_.Map.Key, invalid) || referencesInvalidData(type_.Map.Value, invalid)
	}
	return false
}

func (p *Analysis) checkActorAudiences(audiences []*model.ActorAudience, ownerPos fmt.Stringer, ownerKind string, ownerName string) bool {
	valid := true
	for _, audience := range audiences {
		actor := p.actorByRef(audience.Actor)
		if !p.reporter.checkReference(actor != nil, `%s %s %s references undefined actor "%s"`, ownerPos, ownerKind, ownerName, audience.Actor) {
			valid = false
			continue
		}
		if audience.Via == "" {
			continue
		}
		_, ok := sliceutil.Find(actor.Vias, func(via *model.ActorVia) bool {
			return via.Name == audience.Via
		})
		if !p.reporter.checkReference(ok, `%s %s %s for %s references undefined actor via "%s"`, ownerPos, ownerKind, ownerName, audience.Actor, audience.Via) {
			valid = false
		}
	}
	return valid
}

func (p *Analysis) actorByRef(actorName string) *model.Actor {
	qualifier, name, ok := nameutil.SplitQualified(actorName)
	if !ok {
		return p.actorsMap[actorName]
	}
	import_ := p.importsMap[qualifier]
	if import_ == nil {
		return nil
	}
	return import_.Domain.actorsMap[name]
}

func (p *Analysis) resourceByRef(resourceName string) *model.Resource {
	qualifier, name, ok := nameutil.SplitQualified(resourceName)
	if !ok {
		return p.resourcesMap[resourceName]
	}
	import_ := p.importsMap[qualifier]
	if import_ == nil {
		return nil
	}
	return import_.Domain.resourcesMap[name]
}
