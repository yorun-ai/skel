{{- define "serviceERServer" -}}{{ if not .ClientOnly }}
// {{ .Name }} / ERServer
{{- if .DeprecatedCommentLines }}
//
{{- range .DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ .ERServerName }} interface {
{{- range .Methods }}
	{{ .Name }}(
	{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
	){{ if .ResultType }} ({{ .ResultType.Plain }}, ex.Error){{ else }} ex.Error{{ end }}
{{- end }}

	mustBe{{ .ERServerName }}()
}

// {{ .Name }} / ERServer / WrapperERServer

type {{ .WrapperERServerName }} struct {
	{{ .DefaultServerName }}
	serverImpl {{ .ServerName }}
}

func {{ .WrapperERServerCtorName }}(serverImpl {{ .ServerName }}) {{ .ERServerName }} {
	return &{{ .WrapperERServerName }}{
		serverImpl: serverImpl,
	}
}

func (service *{{ .WrapperERServerName }}) server() {{ .ServerName }} {
	if service.serverImpl == nil {
		return &service.{{ .DefaultServerName }}
	}
	return service.serverImpl
}

{{ range .Methods }}func ({{ .ServerReceiverName }} *{{ $.WrapperERServerName }}) {{ .Name }}(
{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
) ({{ if .ResultType }}{{ .ResultName }} {{ .ResultType.Plain }}, {{ end }}{{ .ErrorName }} ex.Error) {
	defer func() { {{ .ErrorName }} = ex.Recover(recover()) }()
	{{ if .ResultType }}{{ .ResultName }} = {{ end }}{{ .ServerReceiverName }}.server().{{ .Name }}({{ range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }}{{ end }})
	return
}

{{ end -}}

func (*{{ .WrapperERServerName }}) mustBe{{ .ERServerName }}() {}

// {{ .Name }} / ERServer / DefaultERServer

type {{ .DefaultERServerName }} struct {
	{{ .WrapperERServerName }}
}
{{ end }}{{ end }}
