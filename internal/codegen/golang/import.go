package golang

import (
	"fmt"
	"go/token"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	gomodule "go.yorun.ai/skelc/internal/codegen/golang/module"
	"go.yorun.ai/skelc/internal/codegen/golang/view"
	"go.yorun.ai/skelc/internal/model"
)

func (g *_Gen) resolveExternalTypeImports() error {
	types := []*common.ImportBinding{}
	g.bindings = common.TypeBindings{}
	err := g.visitDomainTypes(func(type_ *model.Type) error {
		if type_.ExternalDomain == "" {
			return nil
		}
		path, err := g.goImportPath(type_.ExternalDomain)
		if err != nil {
			return err
		}
		binding := new(common.ImportBinding{Domain: type_.ExternalDomain, Alias: type_.ExternalAlias, Explicit: type_.ExternalAliasExplicit})
		g.bindings[type_] = binding
		types = append(types, binding)
		binding.Path = path
		if !type_.ExternalAliasExplicit {
			binding.Alias = importPackageName(type_.ExternalDomain, true)
			if g.mode == view.ModeApi {
				binding.Alias = importPackageName(type_.ExternalDomain, false) + "api"
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	reserved := []string{"context", "fmt", "errors", "reflect", "sync", "time", "json", "http", "url", "strings", "strconv", "vine", "vrpc", "skel", "meta", "ex", "rpc", "web", "task", "_", "any", "bool", "byte", "error", "int", "string", "float64", "nil", "true", "false"}
	for keyword := token.BREAK; keyword <= token.VAR; keyword++ {
		if keyword.IsKeyword() {
			reserved = append(reserved, keyword.String())
		}
	}
	for _, data := range g.domain.Data() {
		reserved = append(reserved, data.Name)
	}
	for _, enum := range g.domain.Enums() {
		reserved = append(reserved, enum.Name)
	}
	for _, service := range g.domain.Services() {
		for _, method := range service.Methods {
			for _, arg := range method.Arguments {
				reserved = append(reserved, arg.Name)
			}
		}
	}
	common.ResolveImportAliases(types, func(domain string) string { return strings.ReplaceAll(strings.ReplaceAll(domain, ".", ""), "_", "") }, reserved)
	return nil
}

func (g *_Gen) goImportPath(domainName string) (string, error) {
	if path := g.goImports[domainName]; path != "" {
		return gomodule.ImportPath(path)
	}
	if g.modulePrefix == "" {
		return "", fmt.Errorf("missing Go import for domain %s; pass --go-import %s=PACKAGE or --go-module-prefix", domainName, domainName)
	}
	if g.mode == view.ModeApi {
		return buildModuleName(g.modulePrefix, strings.Split(domainName, "."), false) + "api", nil
	}
	return buildModuleName(g.modulePrefix, strings.Split(domainName, "."), true), nil
}

func (g *_Gen) visitDomainTypes(visit common.TypeVisitor) error {
	types := make([]*model.Type, 0)
	if g.mode == view.ModeApi {
		return common.WalkTypes(common.ApiTypeRoots(g.view.Data, g.view.Services), visit)
	}
	for _, dataType := range g.domain.Data() {
		types = appendDataTypes(types, dataType)
	}
	for _, config := range g.domain.Configs() {
		types = appendDataTypes(types, config)
	}
	for _, event := range g.domain.Events() {
		types = appendDataTypes(types, event)
	}
	for _, service := range g.domain.Services() {
		types = appendServiceTypes(types, service)
	}
	for _, actor := range g.domain.Actors() {
		types = appendServiceTypes(types, actor.AuthService)
		types = appendServiceTypes(types, actor.PermService)
	}
	for _, resource := range g.domain.Resources() {
		types = appendServiceTypes(types, resource.CheckService)
	}
	for _, task := range g.domain.Tasks() {
		for _, trigger := range task.Triggers {
			for _, argument := range trigger.Arguments {
				types = append(types, argument.Type)
			}
		}
	}
	return common.WalkTypes(types, visit)
}

func appendServiceTypes(types []*model.Type, service *model.Service) []*model.Type {
	if service == nil {
		return types
	}
	for _, method := range service.Methods {
		types = append(types, method.ResultType)
		for _, argument := range method.Arguments {
			types = append(types, argument.Type)
		}
	}
	return types
}

func appendDataTypes(types []*model.Type, dataType *model.Data) []*model.Type {
	for _, member := range dataType.Members {
		types = append(types, member.Type)
	}
	return types
}
