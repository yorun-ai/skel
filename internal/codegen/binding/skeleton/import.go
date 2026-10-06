package skeleton

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
)

func collectTypeImports(domain *model.Domain, view *codegen.PublicView) []*model.Import {
	used := map[string]struct{}{}
	types := make([]*model.Type, 0)
	for _, data := range view.Data {
		types = appendDataImportTypes(types, data)
	}
	for _, config := range view.Configs {
		types = appendDataImportTypes(types, config)
	}
	for _, resource := range view.Resources {
		for _, check := range resource.Checks {
			for _, argument := range renderResourceCheckArguments(check) {
				types = append(types, argument.Type)
			}
		}
		for _, action := range resource.Actions {
			for _, check := range action.Checks {
				for _, argument := range renderResourceCheckArguments(check) {
					types = append(types, argument.Type)
				}
			}
		}
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectDataImports(domain *model.Domain, dataList []*model.Data) []*model.Import {
	used := map[string]struct{}{}
	types := make([]*model.Type, 0)
	for _, data := range dataList {
		types = appendDataImportTypes(types, data)
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectActorImports(domain *model.Domain, actors []*model.Actor) []*model.Import {
	used := map[string]struct{}{}
	types := make([]*model.Type, 0)
	for _, actor := range actors {
		if actor.AuthEnabled {
			types = appendDataImportTypes(types, actor.AuthCredential)
			types = appendDataImportTypes(types, actor.AuthInfo)
		}
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectServiceImports(domain *model.Domain, services []*model.Service) []*model.Import {
	used := map[string]struct{}{}
	importDomains := make(map[string]string, len(domain.Imports()))
	for _, import_ := range domain.Imports() {
		importDomains[import_.Alias] = import_.Name
	}
	types := make([]*model.Type, 0)
	for _, service := range services {
		for _, audience := range service.Audiences {
			collectImportFromQualifiedName(used, importDomains, audience.Actor)
		}
		for _, method := range service.Methods {
			types = append(types, method.ResultType)
			for _, argument := range method.Arguments {
				types = append(types, argument.Type)
			}
		}
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectImportFromQualifiedName(used map[string]struct{}, importDomains map[string]string, name string) {
	qualifier, _, ok := nameutil.SplitQualified(name)
	if domainName := importDomains[qualifier]; ok && domainName != "" {
		used[domainName] = struct{}{}
	}
}

func appendDataImportTypes(types []*model.Type, data *model.Data) []*model.Type {
	for _, member := range data.Members {
		types = append(types, member.Type)
	}
	return types
}

func collectImportsFromTypes(used map[string]struct{}, types []*model.Type) {
	codegen.VisitTypes(types, func(current *model.Type) {
		if current.ExternalDomain != "" {
			used[current.ExternalDomain] = struct{}{}
		}
	})
}

func selectUsedImports(imports []*model.Import, used map[string]struct{}) []*model.Import {
	selected := make([]*model.Import, 0, len(used))
	for _, import_ := range imports {
		if _, ok := used[import_.Name]; ok {
			selected = append(selected, import_)
		}
	}
	return selected
}
