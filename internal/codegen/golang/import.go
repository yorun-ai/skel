package golang

import (
	"fmt"
	"go/token"
	"strings"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/codegen/golang/view"
	"go.yorun.ai/skel/internal/model"
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
		return ImportPath(path)
	}
	if g.modulePrefix == "" {
		return "", fmt.Errorf("missing Go import for domain %s; pass --go-import %s=PACKAGE or --go-module-prefix", domainName, domainName)
	}
	if g.mode == view.ModeApi {
		return buildModuleName(g.modulePrefix, strings.Split(domainName, "."), false) + "api", nil
	}
	return buildModuleName(g.modulePrefix, strings.Split(domainName, "."), true), nil
}

// usedModuleImports derives dependencies from the types emitted by this view.
func (g *_Gen) usedModuleImports() (map[string]string, error) {
	roots := g.view.TypeRoots()
	imports := map[string]string{}
	for _, domain := range common.ExternalDomains(roots) {
		path := g.goImports[domain]
		if path == "" {
			var err error
			path, err = g.goImportPath(domain)
			if err != nil {
				return nil, err
			}
		}
		imports[domain] = path
	}
	return imports, nil
}

func (g *_Gen) visitDomainTypes(visit common.TypeVisitor) error {
	if g.mode == view.ModeApi {
		return common.WalkTypes(g.view.TypeRoots(), visit)
	}
	// Facade payloads also adapt public declarations, even when their emitted
	// aliases only import the domain's own public package.
	return common.WalkTypes(view.Full(g.domain).TypeRoots(), visit)
}
