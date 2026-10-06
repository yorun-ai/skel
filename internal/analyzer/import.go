package analyzer

import (
	"fmt"
	"maps"
	"slices"

	"go.yorun.ai/skel/schema"
)

func (p *Analysis) skelName(name string) string {
	return fmt.Sprintf("%s.%s", p.name, name)
}

func (p *Analysis) loadImports(domainByName map[string]*Analysis) {
	for _, grammarImport := range p.content.Imports {
		if p.reporter.cancelled() {
			break
		}
		domainName := grammarImport.Domain.String()
		alias := domainName
		if grammarImport.Alias != nil {
			alias = grammarImport.Alias.Value
		}
		importedDomain := domainByName[domainName]
		if importedDomain == nil {
			p.reporter.report(&MissingImportError{Position: position(grammarImport.Pos), Domain: domainName})
			continue
		}
		if previous, exists := p.importsMap[alias]; exists {
			p.reporter.checkDuplicate(previous.Schema.Name == domainName,
				"%s duplicated import alias %s found, already used by %s",
				grammarImport.Pos, alias, previous.Schema.Name)
			continue
		}
		importSchema := &schema.Import{
			Pos:           position(grammarImport.Pos),
			Domain:        importedDomain.Schema(),
			Name:          domainName,
			Alias:         alias,
			ExplicitAlias: grammarImport.Alias != nil,
		}
		p.importsMap[alias] = &_DomainImport{Domain: importedDomain, Schema: importSchema}
		p.imports = append(p.imports, importSchema)
	}
}

func (p *Analysis) checkDuplicated(name string, namePos schema.Position) bool {
	message := `%s duplicated identifier "%s" found, also present at %s`
	valid := true
	if previous := p.enumsMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	if previous := p.dataMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	if previous := p.actorsMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	if previous := p.servicesMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	if previous := p.websMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	if previous := p.tasksMap[name]; previous != nil {
		p.reporter.reportDuplicatef(message, namePos, name, previous.Pos)
		valid = false
	}
	return valid
}

func (p *Analysis) checkDuplicatedResource(name string, namePos schema.Position) bool {
	if previous := p.resourcesMap[name]; previous != nil {
		p.reporter.reportDuplicatef(`%s duplicated resource "%s" found, also present at %s`, namePos, name, previous.Pos)
		return false
	}
	return true
}

func (p *Analysis) checkActorGeneratedNames() {
	generated := map[string]schema.Position{}
	for _, name := range slices.Sorted(maps.Keys(p.actorsMap)) {
		if p.reporter.cancelled() {
			break
		}
		actor := p.actorsMap[name]
		if actor.AuthEnabled {
			p.checkGeneratedIdentifier(actor.AuthCredential.Name, actor.AuthCredential.Pos, generated)
			p.checkGeneratedIdentifier(actor.AuthInfo.Name, actor.AuthInfo.Pos, generated)
			p.checkGeneratedIdentifier(actor.AuthService.Name, actor.Pos, generated)
		}
		if actor.PermissionService != nil {
			p.checkGeneratedIdentifier(actor.PermissionService.Name, actor.Pos, generated)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(p.resourcesMap)) {
		if p.reporter.cancelled() {
			break
		}
		resource := p.resourcesMap[name]
		if resource.CheckService != nil {
			p.checkGeneratedIdentifier(resource.CheckService.Name, resource.Pos, generated)
		}
	}
}

func (p *Analysis) checkGeneratedIdentifier(name string, namePos schema.Position, generated map[string]schema.Position) {
	valid := p.checkDuplicated(name, namePos)
	if previous, duplicated := generated[name]; duplicated {
		p.reporter.reportDuplicatef(`%s duplicated identifier "%s" found, also present at %s`, namePos, name, previous)
		valid = false
	}
	if valid {
		generated[name] = namePos
	}
}

// Retain source import identities even when declarations remain unresolved.
// Querying and comparing a semantic domain must not need a separate alias map.
func (p *Analysis) loadUnresolvedImports() {
	seen := map[string]bool{}
	for _, declaration := range p.content.Imports {
		name := declaration.Domain.String()
		alias := name
		if declaration.Alias != nil {
			alias = declaration.Alias.Value
		}
		if seen[alias] {
			continue
		}
		seen[alias] = true
		p.imports = append(p.imports, new(schema.Import{
			Pos: position(declaration.Pos), Name: name, Alias: alias, ExplicitAlias: declaration.Alias != nil,
		}))
	}
}
