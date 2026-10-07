package analyzer

import (
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
	"maps"
	"slices"
	"sort"
)

func (p *Analysis) finalize() {
	p.schema = nil
	p.enums = slices.Collect(maps.Values(p.enumsMap))
	sort.Slice(p.enums, func(i int, j int) bool {
		return p.enums[i].Name < p.enums[j].Name
	})
	allData := sortData(p.dataMap)
	p.dataList = filterDataByKind(allData, schema.DataKindData)
	p.configs = filterDataByKind(allData, schema.DataKindConfig)
	p.events = filterDataByKind(allData, schema.DataKindEvent)
	p.actors = slices.Collect(maps.Values(p.actorsMap))
	sort.Slice(p.actors, func(i int, j int) bool {
		return p.actors[i].Name < p.actors[j].Name
	})
	p.resources = slices.Collect(maps.Values(p.resourcesMap))
	sort.Slice(p.resources, func(i int, j int) bool {
		return p.resources[i].Name < p.resources[j].Name
	})
	p.services = slices.Collect(maps.Values(p.servicesMap))
	sort.Slice(p.services, func(i int, j int) bool {
		return p.services[i].Name < p.services[j].Name
	})
	p.webs = slices.Collect(maps.Values(p.websMap))
	sort.Slice(p.webs, func(i int, j int) bool {
		return p.webs[i].Name < p.webs[j].Name
	})
	p.tasks = slices.Collect(maps.Values(p.tasksMap))
	sort.Slice(p.tasks, func(i int, j int) bool {
		return p.tasks[i].Name < p.tasks[j].Name
	})
}

func sortData(dataMap map[string]*schema.Data) []*schema.Data {
	dataNames := slices.Sorted(maps.Keys(dataMap))
	return sliceutil.Map(dataNames, func(dataName string) *schema.Data {
		return dataMap[dataName]
	})
}

func filterDataByKind(dataList []*schema.Data, kind schema.DataKind) []*schema.Data {
	return sliceutil.Filter(dataList, func(dataType *schema.Data) bool {
		return dataType.Kind == kind
	})
}
