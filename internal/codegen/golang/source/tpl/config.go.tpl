package {{ $.PackageName }}{{ template "imports" . }}

func init() {
{{- range $s := $.Data }}
	conf.Register({{ $s.SpecName }})
{{- end }}
}
{{ range $s := $.Data }}
var {{ $s.SpecName }} = conf.ConfigSpec{
	Name: "{{ $s.Name }}",
	SkelName: "{{ $s.SkelName }}",
	Hash: "{{ $s.Hash }}",
	Lifecycle: conf.{{ $s.RegisterFunc }},
	Type: reflect.TypeFor[*{{ $s.Name }}](),
}

{{- if $s.CommentLines }}
{{- range $line := $s.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
type {{ $s.Name }} struct {
	conf.ConfigModel
{{- range $sm := $s.Members }}
	{{- if $sm.CommentLines }}
	{{- range $line := $sm.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $sm.Name }} {{ $sm.Type.Plain }} `json:"{{ $sm.SkelName }}"{{ if $sm.Sensitive }} skel:"sensitive"{{ end }}`
{{- end }}
}
{{- if $s.Sensitive }}

func ({{ $s.Name }}) {{ $s.MarkerMethodName }}() {}
{{- end }}

{{ end }}
