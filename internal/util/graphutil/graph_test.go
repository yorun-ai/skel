package graphutil

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestFindCycles(t *testing.T) {
	graph := New[string]()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")
	graph.AddEdge("c", "c")
	graph.AddEdge("d", "e")

	got := sortedCycles(graph.FindCycles())
	want := [][]string{{"a", "b"}, {"c"}}
	if !slices.EqualFunc(got, want, slices.Equal) {
		t.Fatalf("cycles = %v, want %v", got, want)
	}
}

func TestFindCyclesReturnsNothingForAcyclicGraph(t *testing.T) {
	graph := New[string]()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")
	graph.AddEdge("a", "c")

	if cycles := graph.FindCycles(); len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
}

func TestFindCyclesIgnoresNodesWithoutSelfReference(t *testing.T) {
	graph := New[string]()
	graph.AddEdge("a", "b")

	if cycles := graph.FindCycles(); len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
}

// sortedCycles normalizes reported components so the assertions do not depend
// on Tarjan traversal order or on node order inside a component.
func sortedCycles(cycles [][]string) [][]string {
	normalized := make([][]string, 0, len(cycles))
	for _, cycle := range cycles {
		members := slices.Clone(cycle)
		slices.Sort(members)
		normalized = append(normalized, members)
	}
	slices.SortFunc(normalized, slices.Compare)
	return normalized
}

func TestGraphContextCancellation(t *testing.T) {
	graph := New[int]()
	for i := range 1000 {
		graph.AddEdge(i, (i+1)%1000)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	cycles, err := graph.FindCyclesContext(ctx)
	if !errors.Is(err, context.Canceled) || cycles != nil {
		t.Fatalf("unexpected cancelled traversal: %v %v", cycles, err)
	}
	if len(graph.FindCycles()) != 1 {
		t.Fatal("cancellation damaged graph")
	}
}
