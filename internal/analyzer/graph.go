package analyzer

import (
	"context"
	"slices"
)

// _Graph is a directed graph whose nodes preserve insertion order.
type _Graph[N comparable] struct {
	nodes    []N
	edges    map[N][]N
	selfRefs map[N]struct{}
}

// newGraph creates an empty directed graph.
func newGraph[N comparable]() *_Graph[N] {
	return &_Graph[N]{
		nodes:    []N{},
		edges:    map[N][]N{},
		selfRefs: map[N]struct{}{},
	}
}

// addEdge adds a directed edge. Missing nodes are added automatically and
// duplicate edges are ignored.
func (g *_Graph[N]) addEdge(from, to N) {
	g.addNode(from)
	g.addNode(to)
	if slices.Contains(g.edges[from], to) {
		return
	}
	g.edges[from] = append(g.edges[from], to)
	if from == to {
		g.selfRefs[from] = struct{}{}
	}
}

// findCycles returns strongly connected components that contain a cycle.
func (g *_Graph[N]) findCycles() [][]N {
	cycles, _ := g.findCyclesContext(context.Background())
	return cycles
}

// findCyclesContext stops graph traversal when analysis is cancelled.
func (g *_Graph[N]) findCyclesContext(ctx context.Context) ([][]N, error) {
	op := &_TarjanOp[N]{
		ctx:   ctx,
		graph: g.edges,
		nodes: make([]_TarjanNode, 0, len(g.edges)),
		index: make(map[N]int, len(g.edges)),
	}
	for _, node := range g.nodes {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if _, ok := op.index[node]; !ok {
			op.strongConnect(node)
		}
	}

	cycles := make([][]N, 0, len(op.output))
	for _, component := range op.output {
		if len(component) == 1 {
			if _, ok := g.selfRefs[component[0]]; !ok {
				continue
			}
		}
		cycles = append(cycles, component)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return cycles, nil
}

func (g *_Graph[N]) addNode(node N) {
	if _, ok := g.edges[node]; ok {
		return
	}
	g.nodes = append(g.nodes, node)
	g.edges[node] = []N{}
}

type _TarjanOp[N comparable] struct {
	ctx    context.Context
	graph  map[N][]N
	nodes  []_TarjanNode
	stack  []N
	index  map[N]int
	output [][]N
}

type _TarjanNode struct {
	lowLink int
	stacked bool
}

func (op *_TarjanOp[N]) strongConnect(nodeValue N) *_TarjanNode {
	if op.ctx.Err() != nil {
		return nil
	}
	index := len(op.nodes)
	op.index[nodeValue] = index
	op.stack = append(op.stack, nodeValue)
	op.nodes = append(op.nodes, _TarjanNode{lowLink: index, stacked: true})
	node := &op.nodes[index]

	for _, target := range op.graph[nodeValue] {
		if op.ctx.Err() != nil {
			return nil
		}
		targetIndex, seen := op.index[target]
		if !seen {
			targetNode := op.strongConnect(target)
			if targetNode == nil {
				return nil
			}
			if targetNode.lowLink < node.lowLink {
				node.lowLink = targetNode.lowLink
			}
		} else if op.nodes[targetIndex].stacked && targetIndex < node.lowLink {
			node.lowLink = targetIndex
		}
	}

	if node.lowLink == index {
		var component []N
		stackIndex := len(op.stack) - 1
		for {
			target := op.stack[stackIndex]
			targetIndex := op.index[target]
			op.nodes[targetIndex].stacked = false
			component = append(component, target)
			if targetIndex == index {
				break
			}
			stackIndex--
		}
		op.stack = op.stack[:stackIndex]
		op.output = append(op.output, component)
	}

	return node
}
