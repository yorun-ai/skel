{{- define "domainConfigs" -}}
{{- if .Descriptor.Configs }}
	Configs: []*descriptor.Config{
		{{- range $config := .Descriptor.Configs }}
		{{ template "configDescriptorValue" $config }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "configDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ if .Lifecycle }}, Lifecycle: {{ configLifecycleLiteral .Lifecycle }}{{ end }}{{ template "memberDescriptorList" .Members }}}
{{- end }}
