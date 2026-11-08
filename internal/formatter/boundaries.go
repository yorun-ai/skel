package formatter

import (
	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/internal/parser/grammar"
)

// sourceBoundaries locates statements using syntax rather than keyword guesses.
// Invalid fragments retain the token-only layout used by Source.
func sourceBoundaries(source []byte, audiences *[]_Audience) map[int]int {
	content, err := parser.ParseSource("<format>", source)
	if err != nil {
		return nil
	}
	boundaries := make(map[int]int)
	var visit func(any)
	visit = func(node any) {
		switch node := node.(type) {
		case *grammar.SkelContent:
			if node == nil {
				return
			}
			for _, child := range node.Imports {
				visit(child)
			}
			for _, child := range node.Entries {
				visit(child)
			}
			visit(node.Domain)
		case *grammar.DomainContent:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.ImportDecl:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.SkelEntry:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 2)
			visit(node.Enum)
			visit(node.Data)
			visit(node.Config)
			visit(node.Actor)
			visit(node.Resource)
			visit(node.Service)
			visit(node.Web)
			visit(node.Event)
			visit(node.Task)
		case *grammar.Enum:
			if node == nil {
				return
			}
			for _, child := range node.Items {
				visit(child)
			}
		case *grammar.EnumItem:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Data:
			if node == nil {
				return
			}
			for _, child := range node.Members {
				visit(child)
			}
		case *grammar.DataMember:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Event:
			if node == nil {
				return
			}
			visit(node.Payload)
		case *grammar.EventPayload:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Members {
				visit(child)
			}
		case *grammar.Service:
			if node == nil {
				return
			}
			for _, child := range node.Sections {
				visit(child)
			}
		case *grammar.ServiceSection:
			if node == nil {
				return
			}
			for _, child := range node.Decorators {
				visit(child)
			}
			visit(node.Audience)
			visit(node.Auth)
			visit(node.Require)
			visit(node.Method)
		case *grammar.ActorAudience:
			if node == nil {
				return
			}
			audience := _Audience{start: node.Pos.Offset}
			for _, actor := range node.Actors {
				audience.actors = append(audience.actors, actor.Pos.Offset)
				part := actor.Actor.Parts[len(actor.Actor.Parts)-1]
				audience.end = part.Pos.Offset + len(part.Value)
				if actor.Via != nil {
					audience.end = actor.Via.Pos.Offset + len(actor.Via.Value)
				}
				audience.blockStart = 0
				if actor.Block != nil {
					audience.blockStart = actor.Block.Pos.Offset
				}

			}
			*audiences = append(*audiences, audience)

			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.AuthMarker:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Require:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Method:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 2)
			visit(node.Auth)
			visit(node.Require)
			visit(node.Input)
			visit(node.Output)
		case *grammar.MethodInput:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Arguments {
				visit(child)
			}
		case *grammar.MethodOutput:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Argument:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Web:
			if node == nil {
				return
			}
			for _, child := range node.Sections {
				visit(child)
			}
		case *grammar.WebSection:
			if node == nil {
				return
			}
			visit(node.Audience)
			visit(node.Auth)
			visit(node.Mount)
		case *grammar.WebMount:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Actor:
			if node == nil {
				return
			}
			for _, child := range node.Sections {
				visit(child)
			}
		case *grammar.ActorSection:
			if node == nil {
				return
			}
			visit(node.Via)
			visit(node.Auth)
			visit(node.Permission)
		case *grammar.ActorVia:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.ActorAuth:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Sections {
				visit(child)
			}
		case *grammar.ActorAuthSection:
			if node == nil {
				return
			}
			for _, child := range node.Decorators {
				visit(child)
			}
			visit(node.Credential)
			visit(node.Info)
		case *grammar.ActorCredential:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Members {
				visit(child)
			}
		case *grammar.ActorInfo:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Members {
				visit(child)
			}
		case *grammar.ActorPermission:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		case *grammar.Resource:
			if node == nil {
				return
			}
			for _, child := range node.Sections {
				visit(child)
			}
		case *grammar.ResourceSection:
			if node == nil {
				return
			}
			for _, child := range node.Decorators {
				visit(child)
			}
			visit(node.Check)
			visit(node.Action)
		case *grammar.ResourceAction:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			for _, child := range node.Checks {
				visit(child)
			}
		case *grammar.ResourceCheck:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
			visit(node.Input)
		case *grammar.Task:
			if node == nil {
				return
			}
			for _, child := range node.Triggers {
				visit(child)
			}
		case *grammar.TaskTrigger:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 2)
			visit(node.Input)
		case *grammar.Decorator:
			if node == nil {
				return
			}
			boundaries[node.Pos.Offset] = max(boundaries[node.Pos.Offset], 1)
		}
	}
	visit(content)
	if content.Domain != nil && len(content.Imports) > 0 {
		boundaries[content.Imports[0].Pos.Offset] = 2
	}
	return boundaries
}
