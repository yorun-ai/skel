package common

import (
	"cmp"
	"slices"
	"strconv"
)

type _ImportAliasGroup struct {
	domain    string
	preferred string
	explicit  bool
	types     []*ImportBinding
}

// ResolveImportAliases keeps unambiguous preferred names and allocates stable
// alternatives using the target language's naming convention.
func ResolveImportAliases(types []*ImportBinding, fallback func(string) string, reserved []string) {
	groups := map[string]*_ImportAliasGroup{}
	for _, kind := range types {
		key := kind.Domain + "\x00" + kind.Alias
		group := groups[key]
		if group == nil {
			group = &_ImportAliasGroup{domain: kind.Domain, preferred: kind.Alias}
			groups[key] = group
		}
		group.explicit = group.explicit || kind.Explicit
		group.types = append(group.types, kind)
	}
	ordered := make([]*_ImportAliasGroup, 0, len(groups))
	counts := map[string]int{}
	for _, group := range groups {
		ordered = append(ordered, group)
		counts[group.preferred]++
	}
	slices.SortFunc(ordered, func(a, b *_ImportAliasGroup) int {
		if a.explicit != b.explicit {
			if a.explicit {
				return -1
			}
			return 1
		}
		if order := cmp.Compare(a.domain, b.domain); order != 0 {
			return order
		}
		return cmp.Compare(a.preferred, b.preferred)
	})
	used := map[string]bool{}
	for _, name := range reserved {
		used[name] = true
	}
	pending := make([]*_ImportAliasGroup, 0)
	for _, group := range ordered {
		if !used[group.preferred] && (group.explicit || counts[group.preferred] == 1) {
			used[group.preferred] = true
		} else {
			pending = append(pending, group)
		}
	}
	for _, group := range pending {
		base := fallback(group.domain)
		name := base
		for suffix := 2; used[name]; suffix++ {
			name = base + strconv.Itoa(suffix)
		}
		used[name] = true
		for _, kind := range group.types {
			kind.Alias = name
		}
	}
}
