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

{{ range .Methods }}func (service *{{ $.WrapperERServerName }}) {{ .Name }}(
{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
) ({{ if .ResultType }}ret {{ .ResultType.Plain }}, {{ end }}err ex.Error) {
	defer func() { err = ex.Recover(recover()) }()
	{{ if .ResultType }}ret = {{ end }}service.server().{{ .Name }}({{ range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }}{{ end }})
	return
}

{{ end -}}

func (*{{ .WrapperERServerName }}) mustBe{{ .ERServerName }}() {}

// {{ .Name }} / ERServer / DefaultERServer

type {{ .DefaultERServerName }} struct {
	{{ .WrapperERServerName }}
}
{{ end }}{{ end }}
