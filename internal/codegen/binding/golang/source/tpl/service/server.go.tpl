{{- define "serviceServer" -}}{{ if not .ClientOnly }}
// {{ .Name }} / Server
{{- if .DeprecatedCommentLines }}
//
{{- range .DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ .ServerName }} interface {
{{- range .Methods }}
	{{- if .CommentLines }}
	{{- range $lineIndex, $line := .CommentLines }}
	{{- if eq $lineIndex 0 }}
	// {{ $line }}
	{{- else }}
	//   {{ $line }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{ .Name }}(
	{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
	){{ if .ResultType }} {{ .ResultType.Plain }}{{ end }}
{{- end }}

	mustBe{{ .ServerName }}()
}

// {{ .Name }} / Server / DefaultServer

type {{ .DefaultServerName }} struct{}

{{ range .Methods }}func (*{{ $.DefaultServerName }}) {{ .Name }}(
{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Type.Plain }}{{ end -}}
) {{ if .ResultType }}{{ .ResultType.Plain }} {{ end }}{
	ex.PanicNew(ex.InvalidRequest, "method {{ .SkelName }} is not implemented"){{ if .ResultType }}
	return {{ .ResultType.DefaultValue }}{{ end }}
}

{{ end }}func (*{{ .DefaultServerName }}) mustBe{{ .ServerName }}() {}
{{ end }}{{- end -}}
