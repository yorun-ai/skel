{{- define "domainConfigs" -}}
{{- if .Schema.Configs }}
	Configs: []*skel.ConfigSchema{
		{{- range $config := .Schema.Configs }}
		{{ template "configSchemaValue" $config }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "configSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ if .Lifecycle }}, Lifecycle: {{ quote .Lifecycle }}{{ end }}{{ template "memberSchemaList" .Members }}}
{{- end }}
