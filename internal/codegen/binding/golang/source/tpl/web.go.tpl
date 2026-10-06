package {{ $.PackageName }}{{ template "imports" . }}

func init() {
{{- range $web := $.Webs }}
	web.Register({{ $web.SpecName }})
{{- end }}
}
{{ range $web := $.Webs }}
var {{ $web.SpecName }} = &web.WebSpec{
	Name: "{{ $web.Name }}",
	SkelName: "{{ $web.SkelName }}",
	Hash: "{{ $web.Hash }}",
	{{- if $web.MountPath }}
	MountPath: {{ printf "%q" $web.MountPath }},
	{{- end }}
	ServerType: reflect.TypeFor[{{ $web.ServerName }}](),
	DefaultServerType: reflect.TypeFor[*{{ $web.DefaultServerName }}](),
}

{{- if $web.CommentLines }}
{{- range $line := $web.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
type {{ $web.ServerName }} interface {
	web.Handler

	mustBe{{ $web.ServerName }}()
}

type {{ $web.DefaultServerName }} struct {
}

func (*{{ $web.DefaultServerName }}) Routes(*web.Router) {
	panic("method routes is not implemented")
}

func (*{{ $web.DefaultServerName }}) mustBe{{ $web.ServerName }}() {}
{{ end }}
