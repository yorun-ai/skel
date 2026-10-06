// Package symbol provides declaration identities, lexical scopes and reference
// resolution shared by semantic analysis and language tooling. It accepts
// recovered syntax and does not require a valid semantic model.
package symbol

import (
	"strings"

	"go.yorun.ai/skel/internal/source"
)

type Kind uint8

const (
	Unknown Kind = iota
	Enum
	Data
	Config
	Event
	Actor
	Resource
	Service
	Web
	Task
	Parameter
)

type ScopeID string

type SymbolID struct {
	Domain, Name string
	Scope        ScopeID
}

func (id SymbolID) Key() string {
	key := id.Domain + "." + id.Name
	if id.Scope != "" {
		key += "#" + string(id.Scope)
	}
	return key
}
func ParseKey(key string) SymbolID {
	base, scope, _ := strings.Cut(key, "#")
	split := strings.LastIndex(base, ".")
	if split < 0 {
		return SymbolID{}
	}
	return SymbolID{Domain: base[:split], Name: base[split+1:], Scope: ScopeID(scope)}
}

type Symbol struct {
	ID     SymbolID
	Kind   Kind
	Source source.ID
	Span   source.Span
}
type Scope struct {
	ID         ScopeID
	Span       source.Span
	Parameters []Symbol
}
type Reference struct {
	Domain, Name, Qualifier string
	Scope                   ScopeID
	Span                    source.Span
	// Want is Unknown for a value type; actor/resource references specify a kind.
	Want Kind
}

type Status uint8

const (
	Missing Status = iota
	Resolved
	Ambiguous
	UnknownImport
)

type Resolution struct {
	Reference Reference
	Target    SymbolID
	Kind      Kind
	Status    Status
}
type Lookup func(SymbolID) []Symbol

// Resolve preserves Skel's existing precedence: domain enums and data types
// precede lexical generic parameters. Visibility and value-type restrictions
// are subsequent semantic validations, separate from identity binding.
func Resolve(ref Reference, imports map[string]string, lookup Lookup) Resolution {
	result := Resolution{Reference: ref, Target: SymbolID{Domain: ref.Domain, Name: ref.Name}}
	if ref.Qualifier != "" {
		domain, ok := imports[ref.Qualifier]
		if !ok {
			result.Status = UnknownImport
			return result
		}
		result.Target.Domain = domain
	}
	candidates := selectCandidates(lookup(result.Target), ref.Want)
	if len(candidates) == 0 && ref.Qualifier == "" && ref.Scope != "" && ref.Want == Unknown {
		scoped := result.Target
		scoped.Scope = ref.Scope
		candidates = selectCandidates(lookup(scoped), Unknown)
	}
	if len(candidates) == 0 {
		return result
	}
	result.Target = candidates[0].ID
	result.Kind = candidates[0].Kind
	result.Status = Resolved
	if len(candidates) > 1 {
		result.Status = Ambiguous
	}
	return result
}
func selectCandidates(symbols []Symbol, want Kind) []Symbol {
	candidates := []Symbol{}
	rank := 100
	for _, symbol := range symbols {
		priority := 100
		if want != Unknown {
			if symbol.Kind == want {
				priority = 0
			}
		} else {
			switch symbol.Kind {
			case Enum:
				priority = 0
			case Data, Config, Event:
				priority = 1
			case Parameter:
				priority = 2
			}
		}
		if priority == 100 || priority > rank {
			continue
		}
		if priority < rank {
			rank = priority
			candidates = nil
		}
		candidates = append(candidates, symbol)
	}
	return candidates
}

type Document struct {
	Domain     string
	Imports    map[string]string
	Symbols    []Symbol
	References []Reference
	Scopes     []Scope
}

func (d *Document) ParametersAt(offset int) []Symbol {
	if d == nil {
		return nil
	}
	for _, scope := range d.Scopes {
		if offset >= scope.Span.Start && offset < scope.Span.End {
			return append([]Symbol{}, scope.Parameters...)
		}
	}
	return nil
}
