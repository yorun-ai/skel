package hasher

import "go.yorun.ai/skel/internal/model"

func FillHashes(domain *model.Domain) error {
	state := newHashState(domain)

	for _, enum := range domain.Enums() {
		enum.Hash = state.enumHash(enum)
	}
	for _, data := range domain.Data() {
		data.Hash = state.dataHash(data)
	}
	for _, config := range domain.Configs() {
		config.Hash = state.dataHash(config)
	}
	for _, web := range domain.Webs() {
		web.Hash = state.webHash(web)
	}
	for _, event := range domain.Events() {
		event.Hash = state.dataHash(event)
	}
	for _, actor := range domain.Actors() {
		if actor.AuthEnabled {
			actor.AuthCredential.Hash = state.dataHash(actor.AuthCredential)
			actor.AuthInfo.Hash = state.dataHash(actor.AuthInfo)
			actor.AuthService.Hash = state.serviceHash(actor.AuthService)
		}
		if actor.PermService != nil {
			actor.PermService.Hash = state.serviceHash(actor.PermService)
		}
		actor.Hash = state.actorHash(actor)
	}
	for _, resource := range domain.Resources() {
		if resource.CheckService != nil {
			resource.CheckService.Hash = state.serviceHash(resource.CheckService)
		}
		resource.Hash = state.resourceHash(resource)
	}
	for _, service := range domain.Services() {
		service.Hash = state.serviceHash(service)
	}
	for _, task := range domain.Tasks() {
		task.Hash = state.taskHash(task)
	}

	domainHash := state.hashValue(_DomainHashValue{
		Domain:      domain.Name(),
		Description: domain.Description(),
		Enums: buildNamedValues(domain.Enums(),
			func(enum *model.Enum) string { return enum.SkelName },
			func(enum *model.Enum) string { return enum.Hash }),
		Data: buildNamedValues(domain.Data(),
			func(data *model.Data) string { return data.SkelName },
			func(data *model.Data) string { return data.Hash }),
		Configs: buildNamedValues(domain.Configs(),
			func(config *model.Data) string { return config.SkelName },
			func(config *model.Data) string { return config.Hash }),
		Webs: buildNamedValues(domain.Webs(),
			func(web *model.Web) string { return web.SkelName },
			func(web *model.Web) string { return web.Hash }),
		Events: buildNamedValues(domain.Events(),
			func(event *model.Data) string { return event.SkelName },
			func(event *model.Data) string { return event.Hash }),
		Actors: buildNamedValues(domain.Actors(),
			func(actor *model.Actor) string { return actor.SkelName },
			func(actor *model.Actor) string { return actor.Hash }),
		Resources: buildNamedValues(domain.Resources(),
			func(resource *model.Resource) string { return resource.SkelName },
			func(resource *model.Resource) string { return resource.Hash }),
		Services: buildNamedValues(domain.Services(),
			func(service *model.Service) string { return service.SkelName },
			func(service *model.Service) string { return service.Hash }),
		Tasks: buildNamedValues(domain.Tasks(),
			func(task *model.Task) string { return task.SkelName },
			func(task *model.Task) string { return task.Hash }),
	})
	if state.err != nil {
		return state.err
	}
	model.SetDomainHash(domain, domainHash)
	return nil
}
