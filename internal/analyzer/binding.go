package analyzer

import (
	"go.yorun.ai/skel/internal/binding"
	"go.yorun.ai/skel/internal/model"
)

// The analyzer supplies its declaration tables to the same resolver used by
// recovered editor syntax. Model mutation and semantic validation stay here.
func (r *_RefContext) bindType(kind *model.Type) binding.Resolution {
	imports := map[string]string{}
	enums, data, parameters := r.enums, r.dataList, r.typeParameters
	if kind.ExternalAlias != "" {
		if imported := r.imports[kind.ExternalAlias]; imported != nil {
			imports[kind.ExternalAlias] = imported.Domain.name
			enums, data, parameters = imported.Domain.enumsMap, imported.Domain.dataMap, nil
		}
	}
	reference := binding.Reference{Name: kind.SkelName, Qualifier: kind.ExternalAlias, Scope: "parameters"}
	return binding.Resolve(reference, imports, func(id binding.SymbolID) []binding.Symbol {
		symbols := []binding.Symbol{}
		if id.Scope != "" {
			if parameters[id.Name] != nil {
				symbols = append(symbols, binding.Symbol{ID: id, Kind: binding.Parameter})
			}
		} else {
			if enums[id.Name] != nil {
				symbols = append(symbols, binding.Symbol{ID: id, Kind: binding.Enum})
			}
			if data[id.Name] != nil {
				symbols = append(symbols, binding.Symbol{ID: id, Kind: binding.Data})
			}
		}
		return symbols
	})
}
