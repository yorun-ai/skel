package schema

import "go.yorun.ai/skel/internal/util/nameutil"

func normalizeReferenceNames(document *Document, domainName string, importAliases map[string]string) {
	for _, declaration := range document.Declarations {
		normalizeDeclarationReferences(declaration, domainName, importAliases)
	}
}

func normalizeDeclarationReferences(declaration *Declaration, domainName string, importAliases map[string]string) {
	if declaration.Data != nil {
		normalizeDataReferences(declaration.Data, domainName, importAliases)
	}
	if declaration.Actor != nil {
		normalizeDataReferences(declaration.Actor.AuthCredential, domainName, importAliases)
		normalizeDataReferences(declaration.Actor.AuthInfo, domainName, importAliases)
	}
	if declaration.Resource != nil {
		normalizeResourceCheckReferences(declaration.Resource.Checks, domainName, importAliases)
		for _, action := range declaration.Resource.Actions {
			normalizeResourceCheckReferences(action.Checks, domainName, importAliases)
		}
	}
	if declaration.Service != nil {
		normalizeRequirementReferences(declaration.Service.Require, domainName, importAliases)
		for _, method := range declaration.Service.Methods {
			normalizeArgumentReferences(method.Arguments, domainName, importAliases)
			normalizeTypeReference(method.Result, domainName, importAliases)
			normalizeRequirementReferences(method.Require, domainName, importAliases)
		}
	}
	if declaration.Task != nil {
		for _, trigger := range declaration.Task.Triggers {
			normalizeArgumentReferences(trigger.Arguments, domainName, importAliases)
		}
	}
}

func normalizeDataReferences(value *DataSchema, domainName string, importAliases map[string]string) {
	if value == nil {
		return
	}
	for _, member := range value.Members {
		normalizeTypeReference(member.Type, domainName, importAliases)
	}
}

func normalizeResourceCheckReferences(values []*ResourceCheck, domainName string, importAliases map[string]string) {
	for _, check := range values {
		normalizeArgumentReferences(check.Arguments, domainName, importAliases)
	}
}

func normalizeArgumentReferences(values []*Argument, domainName string, importAliases map[string]string) {
	for _, argument := range values {
		normalizeTypeReference(argument.Type, domainName, importAliases)
	}
}

func normalizeTypeReference(value *Type, domainName string, importAliases map[string]string) {
	if value == nil {
		return
	}
	if value.Kind == TypeKindImportedReference {
		value.Name = canonicalReferenceName(domainName, importAliases, value.Name)
	}
	for _, argument := range value.Arguments {
		normalizeTypeReference(argument, domainName, importAliases)
	}
	normalizeTypeReference(value.Element, domainName, importAliases)
	normalizeTypeReference(value.Key, domainName, importAliases)
	normalizeTypeReference(value.Value, domainName, importAliases)
}

func normalizeRequirementReferences(value *Requirement, domainName string, importAliases map[string]string) {
	if value == nil {
		return
	}
	if value.Check != nil {
		value.Check.Resource = canonicalReferenceName(domainName, importAliases, value.Check.Resource)
		for _, argument := range value.Check.Arguments {
			normalizeTypeReference(argument.Type, domainName, importAliases)
		}
	}
	for _, child := range value.Children {
		normalizeRequirementReferences(child, domainName, importAliases)
	}
}

func canonicalReferenceName(domainName string, importAliases map[string]string, name string) string {
	if name == "" {
		return ""
	}
	if alias, localName, ok := nameutil.SplitQualified(name); ok {
		if importedDomain := importAliases[alias]; importedDomain != "" {
			return importedDomain + "." + localName
		}
		return name
	}
	return domainName + "." + name
}
