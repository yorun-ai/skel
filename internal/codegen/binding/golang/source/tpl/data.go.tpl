package {{ $.PackageName }}{{ template "imports" . }}
{{ range $s := $.Data }}
{{- if $s.CommentLines }}
{{- range $line := $s.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
{{ if not $s.Members -}}
type {{ $s.FullName }} struct{}
{{- else -}}
type {{ $s.FullName }} struct {
	{{- range $sm := $s.Members }}
	{{- if $sm.CommentLines }}
	{{- range $line := $sm.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $sm.Name }} {{ $sm.Type.Plain }} `json:"{{ $sm.SkelName }}"{{ if $sm.Sensitive }} skel:"sensitive"{{ end }}`{{ end }}
}
{{- end }}
{{- if $s.Sensitive }}

func ({{ $s.ReceiverType }}) {{ $s.MarkerMethodName }}() {}
{{- end }}
{{ end }}
