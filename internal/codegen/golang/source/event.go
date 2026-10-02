package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/skelmeta"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

const eventGoFilename = "event.go"

var eventImports = []*Import{
	{Path: "reflect"},
	{Path: "go.yorun.ai/vine/core/event"},
}

var eventGoTemplate = loadGoTemplate("event.go.tpl")

type EventGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Events        []*Event
}

type Event struct {
	Name                      string
	SkelName                  string
	Hash                      string
	SpecName                  string
	CommentLines              []string
	Sensitive                 bool
	MarkerMethodName          string
	ListenerOnly              bool
	EmitterOnly               bool
	EmitterName               string
	EmitterImplName           string
	EmitterCtorName           string
	EmitterMethodName         string
	ListenerName              string
	DefaultListenerName       string
	ERListenerName            string
	WrapperERListenerName     string
	WrapperERListenerCtorName string
	DefaultERListenerName     string
	ListenerMethodName        string
	Members                   []*DataMember
}

func (g *_Gen) genEventGo() {
	payload := g.buildEventGoPayload()
	if len(payload.Events) == 0 {
		return
	}

	g.renderGo(eventGoFilename, eventGoTemplate, payload)
}

func (g *_Gen) buildEventGoPayload() *EventGoPayload {
	payload := &EventGoPayload{
		PackageName: g.pkgName,
		Events:      make([]*Event, 0, len(g.view.Events)),
	}
	for _, tokenEvent := range g.view.Events {
		castedEvent := g.castEvent(tokenEvent, g.eventListenerOnly(tokenEvent), g.eventEmitterOnly(tokenEvent))
		payload.Events = append(payload.Events, castedEvent)
	}

	imports := buildEventImports(payload.Events)
	payload.StdImports, payload.ModuleImports = splitImports(imports)
	return payload
}

func (g *_Gen) eventListenerOnly(event *model.Data) bool {
	return g.isSplitPub() && event.Pub
}

func (g *_Gen) eventEmitterOnly(event *model.Data) bool {
	return g.isSplitRegular() && event.Pub
}

func (g *_Gen) castEvent(p *model.Data, listenerOnly bool, emitterOnly bool) *Event {
	eventName := nameutil.ToCamel(p.Name)
	methodName := strings.TrimSuffix(eventName, "Event")
	event_ := &Event{
		Name:                      eventName,
		SkelName:                  p.SkelName,
		Hash:                      p.Hash,
		SpecName:                  fmt.Sprintf("_%sSpec", eventName),
		CommentLines:              deprecatedGoDocLines(goDocLines(eventName, p.Description), eventName, p.DeprecatedReason),
		Sensitive:                 p.Sensitive,
		MarkerMethodName:          skelmeta.SensitiveMarkerMethodName,
		ListenerOnly:              listenerOnly,
		EmitterOnly:               emitterOnly,
		EmitterName:               fmt.Sprintf("%sEmitter", eventName),
		EmitterImplName:           fmt.Sprintf("_%sEmitter", eventName),
		EmitterCtorName:           fmt.Sprintf("New%sEmitter", eventName),
		EmitterMethodName:         fmt.Sprintf("Emit%s", methodName),
		ListenerName:              fmt.Sprintf("%sListener", eventName),
		DefaultListenerName:       fmt.Sprintf("Default%sListener", eventName),
		ERListenerName:            fmt.Sprintf("%sListenerER", eventName),
		WrapperERListenerName:     fmt.Sprintf("_Wrapper%sListenerER", eventName),
		WrapperERListenerCtorName: fmt.Sprintf("_NewWrapper%sListenerER", eventName),
		DefaultERListenerName:     fmt.Sprintf("Default%sListenerER", eventName),
		ListenerMethodName:        fmt.Sprintf("On%s", methodName),
		Members:                   make([]*DataMember, 0, len(p.Members)),
	}
	for _, member := range p.Members {
		castedMember := g.types.castDataMember(member)
		event_.Members = append(event_.Members, castedMember)
	}
	return event_
}

func buildEventImports(events []*Event) []*Import {
	imports := newImportSet()
	imports.addMany(eventImports)
	for _, event := range events {
		if !event.EmitterOnly {
			imports.add(&Import{Path: "go.yorun.ai/vine/core/ex"})
		}
		for _, member := range event.Members {
			imports.addMany(collectTypeImports(member.Type))
		}
	}
	return imports.sortedValues()
}
