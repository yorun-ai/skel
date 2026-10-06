package skeleton

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

func collectTypeImports(domain *schema.Domain, view *codegen.PublicView) []*schema.Import {
	used := map[string]struct{}{}
	types := make([]*schema.Type, 0)
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

func collectDataImports(domain *schema.Domain, dataList []*schema.Data) []*schema.Import {
	used := map[string]struct{}{}
	types := make([]*schema.Type, 0)
	for _, data := range dataList {
		types = appendDataImportTypes(types, data)
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectActorImports(domain *schema.Domain, actors []*schema.Actor) []*schema.Import {
	used := map[string]struct{}{}
	types := make([]*schema.Type, 0)
	for _, actor := range actors {
		if actor.Auth != nil {
			types = appendDataImportTypes(types, actor.Auth.Credential)
			types = appendDataImportTypes(types, actor.Auth.Info)
		}
	}
	collectImportsFromTypes(used, types)
	return selectUsedImports(domain.Imports(), used)
}

func collectServiceImports(domain *schema.Domain, services []*schema.Service) []*schema.Import {
	used := map[string]struct{}{}
	importDomains := make(map[string]string, len(domain.Imports()))
	for _, import_ := range domain.Imports() {
		importDomains[import_.Alias] = import_.Name
	}
	types := make([]*schema.Type, 0)
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

func appendDataImportTypes(types []*schema.Type, data *schema.Data) []*schema.Type {
	for _, member := range data.Members {
		types = append(types, member.Type)
	}
	return types
}

func collectImportsFromTypes(used map[string]struct{}, types []*schema.Type) {
	codegen.VisitTypes(types, func(current *schema.Type) {
		if current.ExternalDomain != "" {
			used[current.ExternalDomain] = struct{}{}
		}
	})
}

func selectUsedImports(imports []*schema.Import, used map[string]struct{}) []*schema.Import {
	selected := make([]*schema.Import, 0, len(used))
	for _, import_ := range imports {
		if _, ok := used[import_.Name]; ok {
			selected = append(selected, import_)
		}
	}
	return selected
}
