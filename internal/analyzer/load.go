package analyzer

import "go.yorun.ai/skel/internal/parser/grammar"

func (p *Analysis) load() bool {
	if !p.reporter.checkNot(p.content == nil || p.content.Domain == nil || p.content.Domain.Name == nil || p.content.Domain.Name.String() == "",
		"missing domain name") {
		return false
	}
	p.name = p.content.Domain.Name.String()
	p.description = p.content.Domain.Description

	for _, entry := range p.content.Entries {
		if p.reporter.cancelled() {
			break
		}
		if p.reporter.full() {
			break
		}
		loaded := p.loadEntry(entry)
		if !loaded {
			if name := grammarEntryName(entry); name != "" && !p.hasDefinition(name) {
				p.unavailable[name] = true
			}
		}
	}
	p.checkActorGeneratedNames()
	return true
}

func (p *Analysis) loadEntry(entry *grammar.SkelEntry) bool {
	switch {
	case entry.Enum != nil:
		enum, valid := parseEnum(p.reporter, entry.Enum)
		if !valid || !p.checkDuplicated(enum.Name, enum.Pos) {
			return false
		}
		enum.Domain = p.name
		enum.SkelName = p.skelName(enum.Name)
		p.enumsMap[enum.Name] = enum
	case entry.Data != nil:
		data, valid := parseData(p.reporter, entry.Data)
		if !valid || !p.checkDuplicated(data.Name, data.Pos) {
			return false
		}
		data.Domain = p.name
		data.SkelName = p.skelName(data.Name)
		p.dataMap[data.Name] = data
	case entry.Config != nil:
		config, valid := parseConfig(p.reporter, entry.Config)
		if !valid || !p.checkDuplicated(config.Name, config.Pos) {
			return false
		}
		config.Domain = p.name
		config.SkelName = p.skelName(config.Name)
		p.dataMap[config.Name] = config
	case entry.Actor != nil:
		actor, valid := parseActor(p.reporter, entry.Actor)
		if !valid || !p.checkDuplicated(actor.Name, actor.Pos) {
			return false
		}
		actor.SkelName = p.skelName(actor.Name)
		if actor.AuthEnabled {
			actor.AuthCredential.Domain = p.name
			actor.AuthCredential.SkelName = p.skelName(actor.AuthCredential.Name)
			actor.AuthInfo.Domain = p.name
			actor.AuthInfo.SkelName = p.skelName(actor.AuthInfo.Name)
		}
		actor.AuthService = buildActorAuthService(actor)
		actor.PermissionService = buildActorPermissionService(actor)
		p.actorsMap[actor.Name] = actor
	case entry.Resource != nil:
		resource, valid := parseResource(p.reporter, entry.Resource, entry.Pub)
		if !valid || !p.checkDuplicatedResource(resource.Name, resource.Pos) {
			return false
		}
		resource.SkelName = p.skelName(resource.Name)
		for _, action := range resource.Actions {
			if p.reporter.cancelled() {
				break
			}
			action.PermissionCode = resource.SkelName + ":" + action.Name
		}
		resource.CheckService = buildResourceCheckService(p.name, resource)
		p.resourcesMap[resource.Name] = resource
	case entry.Event != nil:
		event, valid := parseEvent(p.reporter, entry.Event)
		if !valid || !p.checkDuplicated(event.Name, event.Pos) {
			return false
		}
		event.Domain = p.name
		event.SkelName = p.skelName(event.Name)
		p.dataMap[event.Name] = event
	case entry.Service != nil:
		service, valid := parseService(p.reporter, entry.Service)
		if !valid || !p.checkDuplicated(service.Name, service.Pos) {
			return false
		}
		service.SkelName = p.skelName(service.Name)
		p.servicesMap[service.Name] = service
	case entry.Web != nil:
		web, valid := parseWeb(p.reporter, entry.Web, entry.Pub)
		if !valid || !p.checkDuplicated(web.Name, web.Pos) {
			return false
		}
		web.SkelName = p.skelName(web.Name)
		p.websMap[web.Name] = web
	case entry.Task != nil:
		valid := p.reporter.checkNot(entry.Pub, "%s task %s does not support pub", entry.Task.Name.Pos, entry.Task.Name.Value)
		task, taskValid := parseTask(p.reporter, entry.Task)
		if !valid || !taskValid || !p.checkDuplicated(task.Name, task.Pos) {
			return false
		}
		task.SkelName = p.skelName(task.Name)
		p.tasksMap[task.Name] = task
	}
	return true
}

func grammarEntryName(entry *grammar.SkelEntry) string {
	switch {
	case entry.Enum != nil:
		return entry.Enum.Name.Value
	case entry.Data != nil:
		return entry.Data.Name.Value
	case entry.Config != nil:
		return entry.Config.Name.Value
	case entry.Actor != nil:
		return entry.Actor.Name.Value
	case entry.Resource != nil:
		return entry.Resource.Name.Value
	case entry.Event != nil:
		return entry.Event.Name.Value
	case entry.Service != nil:
		return entry.Service.Name.Value
	case entry.Web != nil:
		return entry.Web.Name.Value
	case entry.Task != nil:
		return entry.Task.Name.Value
	default:
		return ""
	}
}

func (p *Analysis) hasDefinition(name string) bool {
	return p.enumsMap[name] != nil || p.dataMap[name] != nil || p.actorsMap[name] != nil ||
		p.resourcesMap[name] != nil || p.servicesMap[name] != nil || p.websMap[name] != nil || p.tasksMap[name] != nil
}
