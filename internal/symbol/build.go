package symbol

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/source"
)

// Build indexes only recovered declarations and grammar references. Source
// spans remain byte-based until a protocol adapter converts them.
func Build(revision *source.Document, parsed *grammar.SkelContent) *Document {
	document := new(Document{Imports: map[string]string{}})
	if parsed == nil {
		return document
	}
	if parsed.Domain != nil {
		document.Domain = parsed.Domain.Name.String()
	}
	for _, declaration := range parsed.Imports {
		name := declaration.Domain.String()
		alias := name
		if declaration.Alias != nil {
			alias = declaration.Alias.Value
		}
		document.Imports[alias] = name
	}
	for _, entry := range parsed.Entries {
		name, kind := Declaration(entry)
		if name != nil {
			document.Symbols = append(document.Symbols, Symbol{ID: SymbolID{Domain: document.Domain, Name: name.Value}, Kind: kind, Source: revision.ID(), Span: revision.IdentifierSpan(name.Pos.Line, name.Pos.Column, name.Value)})
		}
	}
	var scope ScopeID
	reference := func(name *grammar.QualifiedName, want Kind) {
		if name == nil || len(name.Parts) == 0 {
			return
		}
		last := name.Parts[len(name.Parts)-1]
		qualifier := []string{}
		for _, part := range name.Parts[:len(name.Parts)-1] {
			qualifier = append(qualifier, part.Value)
		}
		document.References = append(document.References, Reference{Domain: document.Domain, Name: last.Value, Qualifier: strings.Join(qualifier, "."), Scope: scope, Span: revision.IdentifierSpan(last.Pos.Line, last.Pos.Column, last.Value), Want: want})
	}
	var visitType func(*grammar.Type)
	visitType = func(kind *grammar.Type) {
		if kind == nil {
			return
		}
		if kind.Reference != nil {
			reference(kind.Reference.Name, Unknown)
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
			reference(expr.Term.Target.Resource, Resource)
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
	for _, entry := range parsed.Entries {
		scope = ""
		if entry.Data != nil {
			scope = addScope(document, revision, entry.Data)
		}
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
					reference(section.Audience.Actor, Actor)
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
					reference(section.Audience.Actor, Actor)
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
	return document
}

func addScope(document *Document, revision *source.Document, data *grammar.Data) ScopeID {
	if len(data.TypeParameters) == 0 {
		return ""
	}
	id := ScopeID(fmt.Sprintf("%x", sha256.Sum256(fmt.Appendf(nil, "%s:%d:%d", revision.ID(), data.Pos.Line, data.Pos.Column))))
	scope := Scope{ID: id, Span: source.Span{Start: revision.Offset(data.Pos.Line, data.Pos.Column), End: revision.Offset(data.EndPos.Line, data.EndPos.Column)}}
	for _, parameter := range data.TypeParameters {
		if parameter.Name == nil {
			continue
		}
		name := parameter.Name
		symbol := Symbol{ID: SymbolID{Domain: document.Domain, Name: name.Value, Scope: id}, Kind: Parameter, Source: revision.ID(), Span: revision.IdentifierSpan(name.Pos.Line, name.Pos.Column, name.Value)}
		scope.Parameters = append(scope.Parameters, symbol)
		document.Symbols = append(document.Symbols, symbol)
	}
	document.Scopes = append(document.Scopes, scope)
	return id
}

func Declaration(entry *grammar.SkelEntry) (*grammar.Identifier, Kind) {
	switch {
	case entry.Enum != nil:
		return entry.Enum.Name, Enum
	case entry.Data != nil:
		return entry.Data.Name, Data
	case entry.Config != nil:
		return entry.Config.Name, Config
	case entry.Event != nil:
		return entry.Event.Name, Event
	case entry.Actor != nil:
		return entry.Actor.Name, Actor
	case entry.Resource != nil:
		return entry.Resource.Name, Resource
	case entry.Service != nil:
		return entry.Service.Name, Service
	case entry.Web != nil:
		return entry.Web.Name, Web
	case entry.Task != nil:
		return entry.Task.Name, Task
	}
	return nil, Unknown
}
