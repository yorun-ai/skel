package typescript

import (
	"sort"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
)

func (b *_WireSchemaBuilder) collectMethod(method *model.Method) {
	if methodArgumentsContainBinary(method) {
		for _, argument := range method.Arguments {
			b.types = append(b.types, argument.Type)
		}
	}
	if methodResultContainsBinary(method) {
		b.types = append(b.types, method.ResultType)
	}
}

func (b *_WireSchemaBuilder) prepareFactoryNames() {
	codegen.VisitTypeGraphs(b.types, func(current *model.Type) {
		if current.Kind == model.TypeKindData {
			b.data[current.Data] = true
		}
	})
	dataList := b.sortedData()
	nameCounts := map[string]int{}
	for _, dataType := range dataList {
		nameCounts[dataType.Name]++
	}

	for _, dataType := range dataList {
		name := dataType.Name
		if nameCounts[name] > 1 {
			name = dataType.SkelName
			if name == "" {
				name = dataType.Domain + "." + dataType.Name
			}
		}
		b.factoryNames[dataType] = "create" + nameutil.ToCamel(name) + "WireSchema"
	}
}

func (b *_WireSchemaBuilder) sortedData() []*model.Data {
	dataList := make([]*model.Data, 0, len(b.data))
	for dataType := range b.data {
		dataList = append(dataList, dataType)
	}
	sort.Slice(dataList, func(i, j int) bool {
		left := dataList[i].SkelName
		if left == "" {
			left = dataList[i].Domain + "." + dataList[i].Name
		}
		right := dataList[j].SkelName
		if right == "" {
			right = dataList[j].Domain + "." + dataList[j].Name
		}
		return left < right
	})
	return dataList
}
