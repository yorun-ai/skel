package analyzer

import (
	"strings"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/graphutil"
	"go.yorun.ai/skelc/internal/util/sliceutil"
)

func (p *Analysis) checkHardCycleReferences(dataList []*model.Data) {
	graph := graphutil.New[*model.Data]()
	refs := _RefsMatrix{}

	for _, dataType := range dataList {
		if refs.has(dataType) {
			continue
		}

		refs[dataType] = _Refs{}
		for _, member := range dataType.Members {
			refs[dataType].merge(referencedData(member.Type))
		}
		for refData := range refs[dataType] {
			graph.AddEdge(dataType, refData)
		}
	}

	for _, cycle := range graph.FindCycles() {
		cycle = append(cycle, cycle[0])
		isHard := true

		for si, di := 0, 1; di < len(cycle); si, di = si+1, di+1 {
			if !refs.refKind(cycle[si], cycle[di]).isHard() {
				isHard = false
				break
			}
		}

		if isHard {
			names := sliceutil.Map(cycle, func(dataType *model.Data) string { return dataType.Name })
			p.reporter.reportf("%s hard reference chain detected: %s, try nullable/list/map instead", cycle[0].Pos, strings.Join(names, " -> "))
		}
	}
}
