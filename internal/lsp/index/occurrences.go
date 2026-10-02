package index

import (
	"slices"
	"strings"

	"go.lsp.dev/protocol"
	"go.yorun.ai/skelc/internal/lsp/source"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

// Only declarations and grammar references are safe rename targets. The
// tolerant declaration index remains available for incomplete source, while
// references come exclusively from successfully recovered syntax.
func indexOccurrences(document *Document) []Occurrence {
	occurrences := make([]Occurrence, 0, len(document.Definitions))
	seen := map[protocol.Range]bool{}
	add := func(key string, range_ protocol.Range) {
		if key != "" && !seen[range_] {
			seen[range_] = true
			occurrences = append(occurrences, Occurrence{Key: key, Range: range_})
		}
	}
	for _, definition := range document.Definitions {
		if !definition.Confirmed {
			continue
		}
		add(definition.Key, definition.Range)
	}
	reference := func(name *grammar.QualifiedName) {
		if name == nil || len(name.Parts) == 0 {
			return
		}
		last := name.Parts[len(name.Parts)-1]
		key := document.Domain + "." + last.Value
		if len(name.Parts) > 1 {
			parts := make([]string, 0, len(name.Parts)-1)
			for _, part := range name.Parts[:len(name.Parts)-1] {
				parts = append(parts, part.Value)
			}
			domain := document.Imports[strings.Join(parts, ".")]
			if domain == "" {
				return
			}
			key = domain + "." + last.Value
		}
		add(key, identifierRange(document.Buffer, last.Pos, last.Value))
	}
	var visitType func(*grammar.Type)
	visitType = func(kind *grammar.Type) {
		if kind == nil {
			return
		}
		if kind.Reference != nil {
			reference(kind.Reference.Name)
			for _, arg := range kind.Reference.TypeArguments {
				visitType(arg)
			}
		}
		if kind.List != nil {
			visitType(kind.List.Value)
		}
		if kind.Map != nil {
			visitType(kind.Map.Key)
			visitType(kind.Map.Value)
		}
	}
	members := func(values []*grammar.DataMember) {
		for _, member := range values {
			visitType(member.Type)
		}
	}
	input := func(value *grammar.MethodInput) {
		if value != nil {
			for _, arg := range value.Arguments {
				visitType(arg.Type)
			}
		}
	}
	var require func(*grammar.RequireExpr)
	require = func(expr *grammar.RequireExpr) {
		if expr == nil {
			return
		}
		if expr.Term != nil && expr.Term.Target != nil {
			reference(expr.Term.Target.Resource)
		}
		for _, child := range expr.Children {
			require(child)
		}
	}
	check := func(value *grammar.ResourceCheck) {
		if value != nil {
			input(value.Input)
		}
	}
	if document.Parsed != nil {
		for _, entry := range document.Parsed.Entries {
			switch {
			case entry.Data != nil:
				members(entry.Data.Members)
			case entry.Config != nil:
				members(entry.Config.Members)
			case entry.Event != nil:
				if entry.Event.Payload != nil {
					members(entry.Event.Payload.Members)
				}
			case entry.Actor != nil:
				for _, section := range entry.Actor.Sections {
					if section.Auth != nil {
						if section.Auth.Credential != nil {
							members(section.Auth.Credential.Members)
						}
						if section.Auth.Info != nil {
							members(section.Auth.Info.Members)
						}
					}
				}
			case entry.Service != nil:
				for _, section := range entry.Service.Sections {
					if section.Audience != nil {
						reference(section.Audience.Actor)
					}
					if section.Require != nil {
						require(section.Require.Expr)
					}
					if method := section.Method; method != nil {
						input(method.Input)
						if method.Output != nil {
							visitType(method.Output.Type)
						}
						if method.Require != nil {
							require(method.Require.Expr)
						}
					}
				}
			case entry.Web != nil:
				for _, section := range entry.Web.Sections {
					if section.Audience != nil {
						reference(section.Audience.Actor)
					}
				}
			case entry.Resource != nil:
				for _, section := range entry.Resource.Sections {
					check(section.Check)
					if section.Action != nil {
						for _, value := range section.Action.Checks {
							check(value)
						}
					}
				}
			case entry.Task != nil:
				for _, trigger := range entry.Task.Triggers {
					input(trigger.Input)
				}
			}
		}
	}
	slices.SortFunc(occurrences, func(a, b Occurrence) int { return source.ComparePosition(a.Range.Start, b.Range.Start) })
	return occurrences
}
