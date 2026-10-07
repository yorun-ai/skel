{{- define "serviceClient" -}}{{ if not .ServerOnly }}
// {{ .Name }} / Client
{{- if .DeprecatedCommentLines }}
//
{{- range .DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ .ClientName }} interface { {{- range .Methods }}
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
	{{- range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end -}}
	{{ .OptionsName }} ...rpc.InvokeOption){{ if .ResultType }} {{ .ResultType.Plain }}{{ end }}{{ end }}
}

type {{ .ClientImplName }} struct {
	clientER {{ .ERClientName }}
}

func {{ .ClientCtorName }}(clientER {{ .ERClientName }}) {{ .ClientName }} {
	return &{{ .ClientImplName }}{clientER: clientER}
}
{{ range .Methods }}
func ({{ .ReceiverName }} *{{ $.ClientImplName }}) {{ .Name }}({{ range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end -}}
{{ .OptionsName }} ...rpc.InvokeOption){{ if .ResultType }} {{ .ResultType.Plain }}{{ end }} {
	{{ if .ResultType }}{{ .ResultName }}, {{ end }}{{ .ErrorName }} := {{ .ReceiverName }}.clientER.{{ .Name }}({{ range .Arguments }}{{ .Name }}, {{ end }}{{ .OptionsName }}...)
	ex.PanicIfError({{ .ErrorName }}){{ if .ResultType }}
	return {{ .ResultName }}{{ end }}
}
{{ end }}
{{- end }}{{- end -}}
