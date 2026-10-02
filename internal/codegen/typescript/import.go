package typescript

import (
	"fmt"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

func (g *_Gen) resolveExternalTypeImports() {
	types := []*common.ImportBinding{}
	g.bindings = common.TypeBindings{}
	common.VisitTypes(common.ApiTypeRoots(g.apiView.Data, g.apiView.Services), func(type_ *model.Type) {
		if type_ == nil || type_.ExternalDomain == "" {
			return
		}
		binding := new(common.ImportBinding{Domain: type_.ExternalDomain, Alias: type_.ExternalAlias, Explicit: type_.ExternalAliasExplicit})
		g.bindings[type_] = binding
		types = append(types, binding)
		importPath := g.tsImportPath(type_.ExternalDomain)
		binding.Path = importPath
		if !type_.ExternalAliasExplicit {
			binding.Alias = importPackageAlias(type_.ExternalDomain)
		}
	})
	reserved := []string{"VrpcClient", "VrpcRequestOptions", "VrpcWireSchema", "await", "class", "const", "default", "enum", "export", "extends", "function", "import", "interface", "let", "new", "return", "super", "this", "typeof", "var", "void", "yield", "break", "case", "catch", "continue", "debugger", "delete", "do", "else", "false", "finally", "for", "if", "in", "instanceof", "null", "switch", "throw", "true", "try", "while", "with", "implements", "package", "private", "protected", "public", "static"}
	for _, data := range g.apiView.Data {
		reserved = append(reserved, data.Name)
	}
	for _, enum := range g.apiView.Enums {
		reserved = append(reserved, enum.Name)
	}
	for _, service := range g.apiView.Services {
		names := buildServiceNames(service.Name)
		reserved = append(reserved, names.Name, names.SpecName, names.FactoryName)
	}
	common.ResolveImportAliases(types, nameutil.ToLowerCamel, reserved)
}

func (g *_Gen) tsImportPath(domainName string) string {
	if g.err != nil {
		return ""
	}
	if path := g.tsImports[domainName]; path != "" {
		var resolved string
		resolved, g.err = importPath(path)
		return resolved
	}
	if g.moduleScope == "" {
		g.err = fmt.Errorf("missing TypeScript import for domain %s; pass --ts-import %s=PACKAGE or --ts-module-scope", domainName, domainName)
		return ""
	}
	return buildPackageName(g.moduleScope, domainName)
}

func (g *_Gen) resolvedModuleImports() map[string]string {
	imports := make(map[string]string, len(g.domain.Imports()))
	common.VisitTypes(common.ApiTypeRoots(g.apiView.Data, g.apiView.Services), func(kind *model.Type) {
		if kind.ExternalDomain != "" {
			imports[kind.ExternalDomain] = g.tsImportPath(kind.ExternalDomain)
		}
	})
	return imports
}
