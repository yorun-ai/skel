{{ define "serviceInfo" -}}
// {{ .Name }} / Spec

var (
	{{ .SpecName }} = &rpc.ServiceSpec{
		Type: rpc.ServiceSpecType{{ if .ClientOnly }}Client{{ else if .ServerOnly }}Server{{ else }}Both{{ end }},
		Name: "{{ .Name }}",
		SkelName: "{{ .SkelName }}",
		Hash: "{{ .Hash }}",
{{- if .ClientOnly }}
		ClientType: reflect.TypeFor[{{ .ClientName }}](),
		ClientCtor: {{ .ClientCtorName }},
		ERClientType: reflect.TypeFor[{{ .ERClientName }}](),
		ERClientCtor: {{ .ERClientCtorName }},
{{- else if .ServerOnly }}
		ServerType: reflect.TypeFor[{{ .ServerName }}](),
		DefaultServerType: reflect.TypeFor[*{{ .DefaultServerName }}](),

		ERServerType: reflect.TypeFor[{{ .ERServerName }}](),
		WrapperERServerCtor: {{ .WrapperERServerCtorName }},
		DefaultERServerType: reflect.TypeFor[*{{ .DefaultERServerName }}](),
{{- else }}
		ServerType: reflect.TypeFor[{{ .ServerName }}](),
		DefaultServerType: reflect.TypeFor[*{{ .DefaultServerName }}](),
		ClientType: reflect.TypeFor[{{ .ClientName }}](),
		ClientCtor: {{ .ClientCtorName }},

		ERServerType: reflect.TypeFor[{{ .ERServerName }}](),
		WrapperERServerCtor: {{ .WrapperERServerCtorName }},
		DefaultERServerType: reflect.TypeFor[*{{ .DefaultERServerName }}](),
		ERClientType: reflect.TypeFor[{{ .ERClientName }}](),
		ERClientCtor: {{ .ERClientCtorName }},
{{- end }}
		Methods: []*rpc.MethodSpec{ {{ range .Methods }}
			{{ .SpecName }},{{ end }}
		},
	}
{{- range .Methods }}
	{{ .SpecName }} = &rpc.MethodSpec{
		Name: "{{ .Name }}",
		SkelName: "{{ .SkelName }}",
		ArgumentsType: {{ if .ArgumentsData }}reflect.TypeFor[{{ .ArgumentsData.Name }}](){{ else }}nil{{ end }},
		ResultType: {{ if .ResultType }}reflect.TypeFor[{{ .ResultType.Plain }}](){{ else }}nil{{ end }},
		ArgumentsSensitive: {{ if .ArgumentsSensitive }}true{{ else }}false{{ end }},
		ResultSensitive: {{ if .ResultSensitive }}true{{ else }}false{{ end }},
		ArgumentsContainsBinaryType: {{ if .ArgumentsContainsBinaryType }}true{{ else }}false{{ end }},
		ResultContainsBinaryType: {{ if .ResultContainsBinaryType }}true{{ else }}false{{ end }},
		MethodFuncs: []any{
{{- if not $.ServerOnly }}
			{{ $.ClientName }}.{{ .Name }},
			{{ $.ERClientName }}.{{ .Name }},
{{- end }}
{{- if not $.ClientOnly }}
			{{ $.ServerName }}.{{ .Name }},
			{{ $.ERServerName }}.{{ .Name }},
{{- end }}
		},
	}{{ end }}
)
{{- end -}}
