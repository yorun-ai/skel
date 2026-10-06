package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"slices"

	"go.yorun.ai/skel/internal/parser/grammar"
)

// The reverse graph owns invalidation; cached entries contain complete local
// outcomes, including failures. Cancellation is never a cached outcome.
type _WorkspaceGraph struct {
	inputs     map[string]string
	dependents map[string][]string
}
type _CachedWorkspaceResult struct {
	fingerprint string
	diagnostics Diagnostics
	domains     []WorkspaceDomain
	domainCount int
}

func hashSource(h hash.Hash, input Source) {
	fmt.Fprintf(h, "%q %q %q %q %q %t %q\n", input.Path, input.Root, input.Domain, input.ExpectedDomain, input.Document.ID(), input.DirectoryInput, input.Document.AnalysisPath())
	digest := input.Document.Digest()
	h.Write(digest[:])
}
func workspaceFingerprint(inputs []Source) string {
	h := sha256.New()
	for _, input := range inputs {
		hashSource(h, input)
		fmt.Fprintf(h, "%d\n", input.Document.Version())
	}
	return hex.EncodeToString(h.Sum(nil))
}
func (w *WorkspaceAnalyzer) invalidateDomainCache(domains map[string]*_WorkspaceDomain, byName map[string][]*_WorkspaceDomain) {
	next := _WorkspaceGraph{inputs: map[string]string{}, dependents: map[string][]string{}}
	for key, domain := range domains {
		h := sha256.New()
		for _, input := range domain.sources {
			hashSource(h, input)
		}
		fmt.Fprintf(h, "%t %t", domain.invalid, domain.syntaxInvalid)
		var imports []*grammar.ImportDecl
		if domain.merged != nil {
			imports = domain.merged.Imports
		}
		for _, declaration := range imports {
			name := declaration.Domain.String()
			fmt.Fprintf(h, "%q", name)
			if domain.root != "" && !w.options.ResolveIsolatedImports {
				continue
			}
			for _, dependency := range byName[name] {
				fmt.Fprintf(h, "%q", dependency.key)
				next.dependents[dependency.key] = append(next.dependents[dependency.key], key)
			}
		}
		next.inputs[key] = hex.EncodeToString(h.Sum(nil))
	}
	changed := []string{}
	for key, value := range next.inputs {
		if w.graph.inputs[key] != value {
			changed = append(changed, key)
		}
	}
	for key := range w.graph.inputs {
		if _, exists := next.inputs[key]; !exists {
			changed = append(changed, key)
		}
	}
	visited := map[string]bool{}
	for len(changed) > 0 {
		key := changed[0]
		changed = changed[1:]
		if visited[key] {
			continue
		}
		visited[key] = true
		delete(w.domains, key)
		changed = append(changed, w.graph.dependents[key]...)
		changed = append(changed, next.dependents[key]...)
	}
	w.graph = next
}

func cloneDiagnostics(input Diagnostics) Diagnostics {
	result := slices.Clone(input)
	for i := range result {
		result[i].Related = slices.Clone(input[i].Related)
		if input[i].Suggestion != nil {
			result[i].Suggestion = new(*input[i].Suggestion)
		}
	}
	return result
}

func cloneWorkspaceDomains(input []WorkspaceDomain) []WorkspaceDomain {
	result := slices.Clone(input)
	for i := range result {
		result[i].Sources = slices.Clone(input[i].Sources)
	}
	return result
}
