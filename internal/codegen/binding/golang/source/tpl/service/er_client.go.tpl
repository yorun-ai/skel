{{- define "serviceERClient" -}}{{ if not .ServerOnly }}
// {{ .Name }} / ERClient
{{- if .DeprecatedCommentLines }}
//
{{- range .DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ .ERClientName }} interface { {{- range .Methods }}
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
	{{ .OptionsName }} ...rpc.InvokeOption){{ if .ResultType }} ({{ .ResultType.Plain }}, ex.Error){{ else }} ex.Error{{ end }}{{ end }}
}

type {{ .ERClientImplName }} struct {
	rpcClient *rpc.Client
}

func {{ .ERClientCtorName }}(rpcClient *rpc.Client) {{ .ERClientName }} {
	return &{{ .ERClientImplName }}{
		rpcClient: rpcClient,
	}
}
{{ range .Methods }}
func ({{ .ReceiverName }} *{{ $.ERClientImplName }}) {{ .Name }}({{ range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end -}}
{{ .OptionsName }} ...rpc.InvokeOption) {{ if .ResultType }}({{ .ResultType.Plain }}, ex.Error){{ else }}ex.Error{{ end }} {
	{{ if .ResultType }}return {{ .ReceiverName }}.rpcClient.InvokeAs[{{ .ResultType.Plain }}]{{ else }}_, {{ .ErrorName }} := {{ .ReceiverName }}.rpcClient.Invoke{{ end }}({{ .SpecName }}.Info(), {{ if .ArgumentsData }}&{{ .ArgumentsData.Name }}{ {{ range .Arguments }}
		{{ .MemberName }}: {{ .Name }},{{ end }}
	{{ "}" }}{{ else }}nil{{ end }}, {{ .OptionsName }}...){{ if not .ResultType }}
	return {{ .ErrorName }}{{ end }}
}
{{ end }}
{{- end }}{{- end -}}
