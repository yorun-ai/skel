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
{{- if $s.Abs }}
 Abs: true,
{{- end }}
{{- if $s.From }}
 FromSkelName: "{{ $s.FromSkelName }}",
 FromType: reflect.TypeFor[*{{ $s.From.Plain }}](),
{{- end }}
}

{{- if $s.CommentLines }}
{{- range $line := $s.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
{{- if $s.From }}
type {{ $s.Name }} {{ $s.From.Plain }}
{{- else }}
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
{{- end }}
{{- if $s.ConversionMethod }}

// {{ $s.ConversionMethod }} exposes the configuration contract without copying its value.
func (c *{{ $s.Name }}) {{ $s.ConversionMethod }}() *{{ if $s.From }}{{ $s.From.Plain }}{{ else }}{{ $s.Name }}{{ end }} {
 return (*{{ if $s.From }}{{ $s.From.Plain }}{{ else }}{{ $s.Name }}{{ end }})(c)
}
{{- end }}
{{- if $s.Sensitive }}

func ({{ $s.Name }}) {{ $s.MarkerMethodName }}() {}
{{- end }}

{{ end }}
