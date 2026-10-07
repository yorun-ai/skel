package {{ $.PackageName }}{{ template "imports" . }}

func init() { {{ range $event := $.Events }}
	event.Register({{ $event.SpecName }}){{ end }}
}
{{ range $event := $.Events }}
var {{ $event.SpecName }} = &event.EventSpec{
	Type: event.EventSpecType{{ if $event.ListenerOnly }}Listener{{ else if $event.EmitterOnly }}Emitter{{ else }}Both{{ end }},
	Name: "{{ $event.Name }}",
	SkelName: "{{ $event.SkelName }}",
	Hash: "{{ $event.Hash }}",
	EmitterMethodName: "{{ $event.EmitterMethodName }}",
	ListenerMethodName: "{{ $event.ListenerMethodName }}",
	PayloadType: reflect.TypeFor[{{ $event.Name }}](),
{{- if not $event.ListenerOnly }}
	EmitterType: reflect.TypeFor[{{ $event.EmitterName }}](),
	EmitterCtor: {{ $event.EmitterCtorName }},
{{- end }}
{{- if not $event.EmitterOnly }}
	ListenerType: reflect.TypeFor[{{ $event.ListenerName }}](),
	DefaultListenerType: reflect.TypeFor[*{{ $event.DefaultListenerName }}](),
	ERListenerType: reflect.TypeFor[{{ $event.ERListenerName }}](),
	WrapperERListenerCtor: {{ $event.WrapperERListenerCtorName }},
	DefaultERListenerType: reflect.TypeFor[*{{ $event.DefaultERListenerName }}](),
{{- end }}
}

{{ if not $event.PayloadAlias }}
{{- if $event.CommentLines }}
{{- range $line := $event.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
{{ if not $event.Members -}}
type {{ $event.Name }} struct{}
{{- else -}}
type {{ $event.Name }} struct { {{ range $member := $event.Members }}
	{{- if $member.CommentLines }}
	{{- range $line := $member.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $member.Name }} {{ $member.Type.Plain }} `json:"{{ $member.SkelName }}"{{ if $member.Sensitive }} skel:"sensitive"{{ end }}`{{ end }}
}
{{- end }}
{{- if $event.Sensitive }}

func ({{ $event.Name }}) {{ $event.MarkerMethodName }}() {}
{{- end }}

{{ end }}
{{ if not $event.ListenerOnly -}}
type {{ $event.EmitterName }} interface {
	{{ $event.EmitterMethodName }}(event *{{ $event.Name }}, _emOpts ...event.EmitOption)
}

type {{ $event.EmitterImplName }} struct {
	emitter *event.Emitter
}

func {{ $event.EmitterCtorName }}(emitter *event.Emitter) {{ $event.EmitterName }} {
	return &{{ $event.EmitterImplName }}{
		emitter: emitter,
	}
}

func (emitter *{{ $event.EmitterImplName }}) {{ $event.EmitterMethodName }}(event *{{ $event.Name }}, _emOpts ...event.EmitOption) {
	emitter.emitter.Emit({{ $event.SpecName }}.Info(), event, _emOpts...)
}
{{ end }}

{{ if not $event.EmitterOnly -}}
type {{ $event.ListenerName }} interface {
	{{ $event.ListenerMethodName }}(event *{{ $event.Name }})

	mustBe{{ $event.ListenerName }}()
}

type {{ $event.DefaultListenerName }} struct{}

func (*{{ $event.DefaultListenerName }}) {{ $event.ListenerMethodName }}(event *{{ $event.Name }}) {
	ex.PanicNew(ex.InvalidRequest, "event {{ $event.Name }} is not implemented")
}

func (*{{ $event.DefaultListenerName }}) mustBe{{ $event.ListenerName }}() {}

type {{ $event.ERListenerName }} interface {
	{{ $event.ListenerMethodName }}(event *{{ $event.Name }}) ex.Error

	mustBe{{ $event.ERListenerName }}()
}

type {{ $event.WrapperERListenerName }} struct {
	{{ $event.DefaultListenerName }}
	listenerImpl {{ $event.ListenerName }}
}

func {{ $event.WrapperERListenerCtorName }}(listenerImpl {{ $event.ListenerName }}) {{ $event.ERListenerName }} {
	return &{{ $event.WrapperERListenerName }}{
		listenerImpl: listenerImpl,
	}
}

func (listener *{{ $event.WrapperERListenerName }}) listener() {{ $event.ListenerName }} {
	if listener.listenerImpl == nil {
		return &listener.{{ $event.DefaultListenerName }}
	}
	return listener.listenerImpl
}

func (listener *{{ $event.WrapperERListenerName }}) {{ $event.ListenerMethodName }}(event *{{ $event.Name }}) (err ex.Error) {
	defer func() { err = ex.Recover(recover()) }()
	listener.listener().{{ $event.ListenerMethodName }}(event)
	return
}

func (*{{ $event.WrapperERListenerName }}) mustBe{{ $event.ERListenerName }}() {}

type {{ $event.DefaultERListenerName }} struct {
	{{ $event.WrapperERListenerName }}
}
{{ end }}
{{ end }}
