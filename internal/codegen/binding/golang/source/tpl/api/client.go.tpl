{{ define "apiClient" }}{{ $s := . }}
{{ range .CommentLines }}// {{ . }}
{{ end }}type {{ .ClientName }} interface {
    {{ range .Methods }}{{ range $index, $line := .CommentLines }}// {{ if $index }}  {{ end }}{{ $line }}
    {{ end }}{{ .Name }}({{ .ContextName }} context.Context, {{ range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end }}{{ .OptionsName }} ...vrpc.InvokeOption) {{ if .ResultType }}({{ .ResultType.Plain }}, error){{ else }}error{{ end }}
    {{ end }}
}

type {{ .ClientImplName }} struct {
    rpcClient *vrpc.Client
}

// {{ .ClientCtorName }} uses the supplied portal client and its credentials.
func {{ .ClientCtorName }}(rpcClient *vrpc.Client) {{ .ClientName }} {
    return &{{ .ClientImplName }}{rpcClient: rpcClient}
}
{{ range .Methods }}
{{ range .CommentLines }}// {{ . }}
{{ end }}func ({{ .ReceiverName }} *{{ $s.ClientImplName }}) {{ .Name }}({{ .ContextName }} context.Context, {{ range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end }}{{ .OptionsName }} ...vrpc.InvokeOption) {{ if .ResultType }}({{ .ResultType.Plain }}, error){{ else }}error{{ end }} {
    {{ if .ResultType }}{{ .ResultName }}{{ else }}_{{ end }}, _, {{ .ErrorName }} := {{ .ReceiverName }}.rpcClient.InvokeAs[{{ if .ResultType }}{{ .ResultType.Plain }}{{ else }}struct{}{{ end }}]({{ .ContextName }}, _{{ $s.Name }}{{ .Name }}Method, {{ if .ArgumentsData }}&{{ .ArgumentsData.Name }}{
        {{ range .Arguments }}{{ .MemberName }}: {{ .Name }},
        {{ end }}
    }{{ else }}nil{{ end }}, {{ .OptionsName }}...)
    return {{ if .ResultType }}{{ .ResultName }}, {{ end }}{{ .ErrorName }}
}
{{ end }}

{{ end }}
