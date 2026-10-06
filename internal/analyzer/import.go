package analyzer

import (
	"fmt"
	"maps"
	"slices"

	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func (p *Analysis) skelName(name string) string {
	return fmt.Sprintf("%s.%s", p.name, name)
}

// Reserve original domain names before checking aliases, including imports in
// other files. Name conflicts do not depend on import order or resolved sources.
func (p *Analysis) validateImportNames() {
	domains := map[string]*grammar.ImportDecl{}
	for _, declaration := range p.content.Imports {
		if p.reporter.cancelled() {
			return
		}
		name := declaration.Domain.String()
		if domains[name] == nil {
			domains[name] = declaration
		}
	}
	qualifiers := map[string]*grammar.ImportDecl{}
	for _, declaration := range p.content.Imports {
		if p.reporter.full() {
			return
		}
		name := declaration.Domain.String()
		qualifier, pos := name, declaration.Domain.Pos
		if declaration.Alias != nil {
			qualifier, pos = declaration.Alias.Value, declaration.Alias.Pos
			if qualifier == p.name {
				p.reporter.reportDuplicatef("%s import alias %s conflicts with current domain %s declared at %s",
					pos, qualifier, p.name, p.content.Domain.Name.Pos)
				continue
			}
			if original := domains[qualifier]; original != nil && qualifier != name {
				p.reporter.reportDuplicatef("%s import alias %s conflicts with imported domain %s declared at %s",
					pos, qualifier, qualifier, original.Domain.Pos)
				continue
			}
		}
		if previous := qualifiers[qualifier]; previous != nil {
			if previous.Domain.String() != name {
				previousPos := previous.Domain.Pos
				if previous.Alias != nil {
					previousPos = previous.Alias.Pos
				}
				p.reporter.reportDuplicatef("%s duplicated import alias %s found, already used by %s at %s",
					pos, qualifier, previous.Domain.String(), previousPos)
			}
			continue
		}
		qualifiers[qualifier] = declaration
	}
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
		if _, exists := p.importsMap[alias]; exists {
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
		if actor.Auth != nil {
			p.checkGeneratedIdentifier(actor.Auth.Credential.Name, actor.Auth.Credential.Pos, generated)
			p.checkGeneratedIdentifier(actor.Auth.Info.Name, actor.Auth.Info.Pos, generated)
			p.checkGeneratedIdentifier(actor.Auth.Service.Name, actor.Pos, generated)
		}
		if actor.Permission != nil {
			p.checkGeneratedIdentifier(actor.Permission.Service.Name, actor.Pos, generated)
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
