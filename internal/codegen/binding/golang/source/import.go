package source

import (
	"fmt"
	"sort"
)

type Import struct {
	Path  string
	Alias string
}

func (import_ *Import) uniquified() string {
	return fmt.Sprintf("%s %s", import_.Path, import_.Alias)
}

type _ImportSet map[string]*Import

func newImportSet() _ImportSet {
	return _ImportSet{}
}

func (imports _ImportSet) add(import_ *Import) {
	if import_ == nil {
		return
	}
	imports[import_.uniquified()] = import_
}

func (imports _ImportSet) addMany(importList ...[]*Import) {
	for _, importItems := range importList {
		for _, import_ := range importItems {
			imports.add(import_)
		}
	}
}

func (imports _ImportSet) sortedValues() []*Import {
	values := make([]*Import, 0, len(imports))
	for _, import_ := range imports {
		values = append(values, import_)
	}
	sort.Slice(values, func(i int, j int) bool {
		return values[i].Path < values[j].Path
	})
	return values
}

func collectTypeImports(kinds ...*Type) []*Import {
	imports := newImportSet()
	for _, kind := range kinds {
		if kind == nil {
			continue
		}
		imports.addMany(kind.Imports, kind.DefaultImports)
	}
	return imports.sortedValues()
}

func (import_ *Import) clone() *Import {
	if import_ == nil {
		return nil
	}
	return &Import{
		Path:  import_.Path,
		Alias: import_.Alias,
	}
}

func cloneImports(imports []*Import) []*Import {
	clonedImports := make([]*Import, 0, len(imports))
	for _, import_ := range imports {
		clonedImports = append(clonedImports, import_.clone())
	}
	return clonedImports
}
