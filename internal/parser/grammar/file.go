package grammar

import (
	"slices"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var topLevelDeclarationKeywords = []string{
	"enum", "data", "config", "actor", "resource", "service", "web", "event", "task",
}

// TopLevelDeclarationKeywords returns every keyword that starts a Skel declaration.
func TopLevelDeclarationKeywords() []string {
	return slices.Clone(topLevelDeclarationKeywords)
}

// IsTopLevelDeclarationKeyword reports whether value starts a Skel declaration.
func IsTopLevelDeclarationKeyword(value string) bool {
	return slices.Contains(topLevelDeclarationKeywords, value)
}

type SkelContent struct {
	Pos     lexer.Position
	Domain  *DomainContent `parser:"(Newline* @@)?"`
	Imports []*ImportDecl  `parser:"(Newline* @@)*"`
	Entries []*SkelEntry   `parser:"(Newline* @@)* Newline*"`
}

func (content *SkelContent) Finalize() error {
	if content == nil {
		return nil
	}
	for _, entry := range content.Entries {
		if entry.Ext && (entry.Pub || entry.Api || entry.Service == nil) {
			return participle.Errorf(entry.Pos, "ext is only allowed on services and cannot be combined with pub or api")
		}
		if entry.Api && (entry.Pub || entry.Service == nil) {
			return participle.Errorf(entry.Pos, "api is only allowed on services and cannot be combined with pub")
		}
		switch {
		case entry.Enum != nil:
			entry.Enum.Decorators = entry.Decorators
			entry.Enum.Pub = entry.Pub
		case entry.Data != nil:
			entry.Data.Decorators = entry.Decorators
			entry.Data.Pub = entry.Pub
		case entry.Config != nil:
			entry.Config.Decorators = entry.Decorators
			entry.Config.Pub = entry.Pub
		case entry.Actor != nil:
			entry.Actor.Decorators = entry.Decorators
			entry.Actor.Pub = entry.Pub
			if err := entry.Actor.Finalize(); err != nil {
				return err
			}
		case entry.Resource != nil:
			entry.Resource.Decorators = entry.Decorators
			entry.Resource.Finalize()
		case entry.Service != nil:
			entry.Service.Decorators = entry.Decorators
			entry.Service.Pub = entry.Pub
			entry.Service.Api = entry.Api
			entry.Service.Ext = entry.Ext
		case entry.Web != nil:
			entry.Web.Decorators = entry.Decorators
			entry.Web.Finalize()
		case entry.Event != nil:
			entry.Event.Decorators = entry.Decorators
			entry.Event.Pub = entry.Pub
		case entry.Task != nil:
			entry.Task.Decorators = entry.Decorators
		}
	}
	return nil
}

type DomainContent struct {
	Pos         lexer.Position
	Decorators  []*Decorator   `parser:"(@@ (Newline)*)*"`
	Name        *QualifiedName `parser:"\"domain\" @@"`
	Description string
}

type ImportDecl struct {
	Pos    lexer.Position
	Domain *QualifiedName `parser:"\"import\" @@"`
	Alias  *Identifier    `parser:"(\"as\" @@)?"`
}

type QualifiedName struct {
	Pos   lexer.Position
	Parts []*Identifier `parser:"@@ (\".\" @@)*"`
}

func (name *QualifiedName) String() string {
	if name == nil {
		return ""
	}
	partValues := make([]string, 0, len(name.Parts))
	for _, part := range name.Parts {
		partValues = append(partValues, part.Value)
	}
	return strings.Join(partValues, ".")
}

func (content *DomainContent) Finalize() error {
	if content == nil {
		return nil
	}
	for _, decorator := range content.Decorators {
		switch decorator.Name.Value {
		case "desc":
			if content.Description != "" {
				return participle.Errorf(decorator.Name.Pos, "duplicated decorator @desc")
			}
			if decorator.Value == nil {
				return participle.Errorf(decorator.Name.Pos, "decorator @desc requires a string argument")
			}
			description, err := UnquoteDescriptionString(decorator.Value.Raw)
			if err != nil {
				return participle.Errorf(decorator.Name.Pos, "%s", err)
			}
			content.Description = description
		default:
			return participle.Errorf(decorator.Name.Pos, "unexpected decorator @%s", decorator.Name.Value)
		}
	}
	return nil
}

type SkelEntry struct {
	Pos        lexer.Position
	Decorators []*Decorator `parser:"(@@ (Newline)*)*"`
	Pub        bool         `parser:"@\"pub\"?"`
	Api        bool         `parser:"@\"api\"?"`
	Ext        bool         `parser:"@\"ext\"?"`
	Enum       *Enum        `parser:"(\"enum\" @@"`
	Data       *Data        `parser:"| \"data\" @@"`
	Config     *Data        `parser:"| \"config\" @@"`
	Actor      *Actor       `parser:"| \"actor\" @@"`
	Resource   *Resource    `parser:"| \"resource\" @@"`
	Service    *Service     `parser:"| \"service\" @@"`
	Web        *Web         `parser:"| \"web\" @@"`
	Event      *Event       `parser:"| \"event\" @@"`
	Task       *Task        `parser:"| \"task\" @@)"`
}
