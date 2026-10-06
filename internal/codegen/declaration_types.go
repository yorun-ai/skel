package codegen

import "go.yorun.ai/skel/internal/model"

// TypeRoots returns the types used by these declarations.
func (v Declarations) TypeRoots(api bool) []*model.Type {
	if api {
		return ApiTypeRoots(v.Data, v.Services)
	}
	var roots []*model.Type
	addData := func(data *model.Data) {
		if data != nil {
			for _, member := range data.Members {
				roots = append(roots, member.Type)
			}
		}
	}
	addService := func(service *model.Service) {
		if service != nil {
			for _, method := range service.Methods {
				roots = append(roots, method.ResultType)
				for _, argument := range method.Arguments {
					roots = append(roots, argument.Type)
				}
			}
		}
	}
	for _, data := range v.Data {
		addData(data)
	}
	for _, config := range v.Configs {
		addData(config)
	}
	for _, event := range v.Events {
		addData(event)
	}
	for _, service := range v.Services {
		addService(service)
	}
	for _, actor := range v.Actors {
		if actor.AuthEnabled {
			addData(actor.AuthCredential)
			addData(actor.AuthInfo)
			addService(actor.AuthService)
		}
		addService(actor.PermService)
	}
	for _, resource := range v.Resources {
		addService(resource.CheckService)
	}
	for _, task := range v.Tasks {
		for _, trigger := range task.Triggers {
			for _, argument := range trigger.Arguments {
				roots = append(roots, argument.Type)
			}
		}
	}
	return roots
}
