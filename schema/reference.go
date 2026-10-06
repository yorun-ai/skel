package schema

import "strings"

// ReferenceName resolves a local name or import qualifier to its domain name.
// Already qualified names are retained. It never loads imported declarations.
func (d *Domain) ReferenceName(name string) string {
	if name == "" {
		return ""
	}
	if i := strings.LastIndex(name, "."); i >= 0 {
		for _, imported := range d.Imports() {
			if imported.Alias == name[:i] {
				return imported.Name + name[i:]
			}
		}
		return name
	}
	return d.Name() + "." + name
}

// TypeReferenceName identifies a named type without traversing its declaration.
// Unresolved imports use the same fully qualified identity as resolved imports.
func (d *Domain) TypeReferenceName(value *Type) string {
	if value == nil {
		return ""
	}
	if value.Kind == TypeKindUnresolvedReference && value.ExternalAlias != "" {
		if value.ExternalDomain != "" {
			return value.ExternalDomain + "." + value.SkelName
		}
		return d.ReferenceName(value.ExternalAlias + "." + value.SkelName)
	}
	if value.SkelName != "" {
		return d.ReferenceName(value.SkelName)
	}
	if value.Data != nil {
		return d.ReferenceName(value.Data.SkelName)
	}
	if value.Enum != nil {
		return d.ReferenceName(value.Enum.SkelName)
	}
	return ""
}
