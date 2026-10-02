package analyzer

import (
	"context"
	"slices"
	"strings"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/util/graphutil"
)

func (p *Analysis) checkHardCycleReferences(dataList []*model.Data) {
	graph := graphutil.New[*model.Data]()
	edges := map[*model.Data][]*model.Data{}
	for _, data := range dataList {
		if p.reporter.cancelled() {
			return
		}
		if _, seen := edges[data]; seen {
			continue
		}
		edges[data] = nil
		refs := _Refs{}
		for _, member := range data.Members {
			if p.reporter.cancelled() {
				return
			}
			refs.merge(referencedData(member.Type))
		}
		for target, kind := range refs {
			if p.reporter.cancelled() {
				return
			}
			if kind.isHard() {
				edges[data] = append(edges[data], target)
			}
		}
		slices.SortFunc(edges[data], func(a, b *model.Data) int { return strings.Compare(a.Name, b.Name) })
		for _, target := range edges[data] {
			graph.AddEdge(data, target)
		}
	}
	ctx := p.reporter.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	components, err := graph.FindCyclesContext(ctx)
	if err != nil {
		return
	}
	for _, component := range components {
		if p.reporter.cancelled() {
			return
		}
		// A strongly connected component is a set, not an ordered cycle. Follow
		// edges inside it until a node repeats to obtain a real diagnostic path.
		members := map[*model.Data]bool{}
		for _, data := range component {
			members[data] = true
		}
		slices.SortFunc(component, func(a, b *model.Data) int { return strings.Compare(a.Name, b.Name) })
		positions := map[*model.Data]int{}
		path := []*model.Data{}
		current := component[0]
		for {
			if p.reporter.cancelled() {
				return
			}
			if start, ok := positions[current]; ok {
				cycle := append(path[start:], current)
				names := make([]string, len(cycle))
				for i, data := range cycle {
					names[i] = data.Name
				}
				p.reporter.reportf("%s hard reference chain detected: %s, try nullable/list/map instead", current.Pos, strings.Join(names, " -> "))
				break
			}
			positions[current] = len(path)
			path = append(path, current)
			for _, target := range edges[current] {
				if members[target] {
					current = target
					break
				}
			}
		}
	}
}
