package workspace

import (
	"fmt"

	"go.yorun.ai/skel/internal/binding"
)

// ValidateRename resolves the affected workspace against a hypothetical symbol
// table. This catches both direct collisions and capture of generic references,
// using exactly the same scope and precedence rules as normal binding.
func (s Snapshot) ValidateRename(document *Document, key, newName string) error {
	oldID := binding.ParseKey(key)
	if oldID.Name == newName {
		return nil
	}
	oldKey := s.ResolveKey(document, key)
	newID := oldID
	newID.Name = newName
	newKey := s.ResolveKey(document, newID.Key())
	if len(s.symbols[newKey]) > 0 {
		return fmt.Errorf("Skel declaration %s already exists", newID.Key())
	}
	renamed := append([]binding.Symbol{}, s.symbols[oldKey]...)
	for i := range renamed {
		renamed[i].ID = newID
	}
	for _, candidate := range s.ordered {
		lookup := func(id binding.SymbolID) []binding.Symbol { return s.symbols[s.ResolveKey(candidate, id.Key())] }
		changedLookup := func(id binding.SymbolID) []binding.Symbol {
			resolved := s.ResolveKey(candidate, id.Key())
			if resolved == oldKey {
				return nil
			}
			if resolved == newKey {
				return renamed
			}
			return s.symbols[resolved]
		}
		for _, reference := range candidate.Bindings.References {
			before := binding.Resolve(reference, candidate.Bindings.Imports, lookup)
			if before.Status != binding.Resolved {
				continue
			}
			expected := before.Target
			if s.ResolveKey(candidate, before.Target.Key()) == oldKey {
				reference.Name = newName
				expected = newID
			}
			after := binding.Resolve(reference, candidate.Bindings.Imports, changedLookup)
			if after.Status != binding.Resolved || after.Target != expected {
				return fmt.Errorf("Skel name %s conflicts with an existing reference binding", newName)
			}
		}
	}
	return nil
}
