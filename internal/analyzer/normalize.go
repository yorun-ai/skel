package analyzer

import (
	"fmt"
	"maps"
	"slices"

	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

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
		if p.reporter.cancelled() {
			break
		}
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
		if p.reporter.cancelled() {
			break
		}
		actor := p.actorsMap[name]
		p.normalizeDataType(actor.AuthCredential, refs)
		p.normalizeDataType(actor.AuthInfo, refs)
		if actor.PermissionService != nil {
			p.normalizeServiceTypes(actor.PermissionService, refs)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.resourcesMap)) {
		if p.reporter.cancelled() {
			break
		}
		resource := p.resourcesMap[name]
		if resource.CheckService != nil {
			p.normalizeServiceTypes(resource.CheckService, refs)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.servicesMap)) {
		if p.reporter.cancelled() {
			break
		}
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
			if p.reporter.cancelled() {
				break
			}
			web := p.websMap[name]
			p.checkActorAudiences(web.Audiences, web.Pos, "web", web.Name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.tasksMap)) {
		if p.reporter.cancelled() {
			break
		}
		task := p.tasksMap[name]
		for _, trigger := range task.Triggers {
			if p.reporter.cancelled() {
				break
			}
			for _, arg := range trigger.Arguments {
				if p.reporter.cancelled() {
					break
				}
				fixTypeRef(p.reporter, arg.Type, refs)
			}
		}
	}
}

func (p *Analysis) validateNormalizedData() {
	allData := sliceutil.Filter(sortData(p.dataMap), func(dataType *schema.Data) bool {
		return !p.invalidData[dataType]
	})
	for _, name := range slices.Sorted(maps.Keys(p.actorsMap)) {
		if p.reporter.cancelled() {
			break
		}
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

func (p *Analysis) normalizeDataType(dataType *schema.Data, refs *_RefContext) bool {
	if dataType == nil {
		return true
	}
	refs.typeParameters = sliceutil.MapToMap(dataType.TypeParameters, func(typeParam *schema.TypeParameter) (string, *schema.TypeParameter) {
		return typeParam.Name, typeParam
	})
	defer func() {
		refs.typeParameters = nil
	}()

	valid := true
	for _, member := range dataType.Members {
		if p.reporter.cancelled() {
			break
		}
		valid = fixTypeRef(p.reporter, member.Type, refs) && valid
	}
	return valid
}

func (p *Analysis) propagateInvalidData() {
	for changed := true; changed; {
		if p.reporter.cancelled() {
			break
		}
		changed = false
		for _, dataType := range p.dataMap {
			if p.reporter.cancelled() {
				break
			}
			if p.invalidData[dataType] {
				continue
			}
			for _, member := range dataType.Members {
				if p.reporter.cancelled() {
					break
				}
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

func referencesInvalidData(type_ *schema.Type, invalid map[*schema.Data]bool) bool {
	if type_ == nil {
		return false
	}
	switch type_.Kind {
	case schema.TypeKindData:
		if invalid[type_.Data] {
			return true
		}
		for _, argument := range type_.TypeArguments {
			if referencesInvalidData(argument, invalid) {
				return true
			}
		}
	case schema.TypeKindList:
		return referencesInvalidData(type_.List.Value, invalid)
	case schema.TypeKindMap:
		return referencesInvalidData(type_.Map.Key, invalid) || referencesInvalidData(type_.Map.Value, invalid)
	}
	return false
}

func (p *Analysis) checkActorAudiences(audiences []*schema.ActorAudience, ownerPos fmt.Stringer, ownerKind string, ownerName string) bool {
	valid := true
	for _, audience := range audiences {
		if p.reporter.cancelled() {
			break
		}
		actor := p.actorByRef(audience.Actor)
		if !p.reporter.checkReference(actor != nil, `%s %s %s references undefined actor "%s"`, ownerPos, ownerKind, ownerName, audience.Actor) {
			valid = false
			continue
		}
		if audience.Via == "" {
			continue
		}
		_, ok := sliceutil.Find(actor.Vias, func(via *schema.ActorVia) bool {
			return via.Name == audience.Via
		})
		if !p.reporter.checkReference(ok, `%s %s %s for %s references undefined actor via "%s"`, ownerPos, ownerKind, ownerName, audience.Actor, audience.Via) {
			valid = false
		}
	}
	return valid
}

func (p *Analysis) actorByRef(actorName string) *schema.Actor {
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

func (p *Analysis) resourceByRef(resourceName string) *schema.Resource {
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
