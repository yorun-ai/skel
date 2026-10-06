package {{ $.PackageName }}{{ template "imports" . }}

{{ if or $.Actors $.AuthServices -}}
func init() { {{ range $actor := $.Actors }}
	meta.RegisterActor(meta.ActorSpec{
		Name: "{{ $actor.Name }}",
		SkelName: "{{ $actor.SkelName }}",
		Hash: "{{ $actor.Hash }}",{{ if $actor.HasInfo }}
		InfoSkelName: "{{ $actor.AuthInfoSkelName }}",
		InfoType: reflect.TypeFor[*{{ $actor.AuthInfoName }}](),{{ end }}
	}){{ end }}{{ range $service := $.AuthServices }}
	rpc.Register({{ $service.SpecName }}){{ end }}
}

{{ end -}}
{{ range $actor := $.Actors }}
{{- if $actor.CommentLines }}
{{- range $line := $actor.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
type {{ $actor.Name }} struct {
	skel.ActorBase
}

func ({{ $actor.Name }}) Name() string {
	return "{{ $actor.Name }}"
}

func ({{ $actor.Name }}) SkelName() string {
	return "{{ $actor.SkelName }}"
}

func ({{ $actor.Name }}) Vias() []skel.ActorVia {
	return []skel.ActorVia{
		{{- range $via := $actor.Vias }}
		{{ $via }},
		{{- end }}
	}
}
{{ end }}
{{ range $s := $.CredentialData }}
{{- if $s.CommentLines }}
{{- range $line := $s.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
{{ if not $s.Members -}}
type {{ $s.FullName }} struct{}
{{- else -}}
type {{ $s.FullName }} struct { {{ range $sm := $s.Members }}
	{{- if $sm.CommentLines }}
	{{- range $line := $sm.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $sm.Name }} {{ $sm.Type.Plain }} `json:"{{ $sm.SkelName }}"{{ if or $sm.Sensitive $sm.Identifier }} skel:"{{ if $sm.Sensitive }}sensitive{{ if $sm.Identifier }},{{ end }}{{ end }}{{ if $sm.Identifier }}identifier{{ end }}"{{ end }}`{{ end }}
}
{{- end }}
{{- if $s.Sensitive }}

func ({{ $s.ReceiverType }}) {{ $s.MarkerMethodName }}() {}
{{- end }}
{{ end }}
{{ range $service := $.AuthServices }}
{{- if $service.CommentLines }}
{{- range $line := $service.CommentLines }}
// {{ $line }}
{{- end }}
{{- else }}
// {{ $service.ServerName }}
{{- end }}

{{ template "serviceInfo" $service -}}
{{ template "serviceArguments" $service -}}
{{ template "serviceServer" $service -}}
{{ template "serviceERServer" $service -}}
{{ template "serviceClient" $service -}}
{{ template "serviceERClient" $service }}
{{- end }}
