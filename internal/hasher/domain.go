package hasher

import "go.yorun.ai/skel/schema"

func FillHashes(domain *schema.Domain) error {
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
		if actor.PermissionService != nil {
			actor.PermissionService.Hash = state.serviceHash(actor.PermissionService)
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
			func(enum *schema.Enum) string { return enum.SkelName },
			func(enum *schema.Enum) string { return enum.Hash }),
		Data: buildNamedValues(domain.Data(),
			func(data *schema.Data) string { return data.SkelName },
			func(data *schema.Data) string { return data.Hash }),
		Configs: buildNamedValues(domain.Configs(),
			func(config *schema.Data) string { return config.SkelName },
			func(config *schema.Data) string { return config.Hash }),
		Webs: buildNamedValues(domain.Webs(),
			func(web *schema.Web) string { return web.SkelName },
			func(web *schema.Web) string { return web.Hash }),
		Events: buildNamedValues(domain.Events(),
			func(event *schema.Data) string { return event.SkelName },
			func(event *schema.Data) string { return event.Hash }),
		Actors: buildNamedValues(domain.Actors(),
			func(actor *schema.Actor) string { return actor.SkelName },
			func(actor *schema.Actor) string { return actor.Hash }),
		Resources: buildNamedValues(domain.Resources(),
			func(resource *schema.Resource) string { return resource.SkelName },
			func(resource *schema.Resource) string { return resource.Hash }),
		Services: buildNamedValues(domain.Services(),
			func(service *schema.Service) string { return service.SkelName },
			func(service *schema.Service) string { return service.Hash }),
		Tasks: buildNamedValues(domain.Tasks(),
			func(task *schema.Task) string { return task.SkelName },
			func(task *schema.Task) string { return task.Hash }),
	})
	if state.err != nil {
		return state.err
	}
	*domain = *schema.NewDomainFromSpec(schema.DomainSpec{
		Name: domain.Name(), Description: domain.Description(), Hash: domainHash,
		Imports: domain.Imports(), Enums: domain.Enums(), Data: domain.Data(),
		Configs: domain.Configs(), Events: domain.Events(), Actors: domain.Actors(),
		Resources: domain.Resources(), Webs: domain.Webs(), Services: domain.Services(),
		Tasks: domain.Tasks(),
	})
	return nil
}
