package analyzer

import (
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/symbol"
)

// The analyzer supplies its declaration tables to the same resolver used by
// recovered editor syntax. Model mutation and semantic validation stay here.
func (r *_RefContext) bindType(kind *model.Type) symbol.Resolution {
	imports := map[string]string{}
	enums, data, parameters := r.enums, r.dataList, r.typeParameters
	if kind.ExternalAlias != "" {
		if imported := r.imports[kind.ExternalAlias]; imported != nil {
			imports[kind.ExternalAlias] = imported.Domain.name
			enums, data, parameters = imported.Domain.enumsMap, imported.Domain.dataMap, nil
		}
	}
	reference := symbol.Reference{Name: kind.SkelName, Qualifier: kind.ExternalAlias, Scope: "parameters"}
	return symbol.Resolve(reference, imports, func(id symbol.SymbolID) []symbol.Symbol {
		symbols := []symbol.Symbol{}
		if id.Scope != "" {
			if parameters[id.Name] != nil {
				symbols = append(symbols, symbol.Symbol{ID: id, Kind: symbol.Parameter})
			}
		} else {
			if enums[id.Name] != nil {
				symbols = append(symbols, symbol.Symbol{ID: id, Kind: symbol.Enum})
			}
			if data[id.Name] != nil {
				symbols = append(symbols, symbol.Symbol{ID: id, Kind: symbol.Data})
			}
		}
		return symbols
	})
}
