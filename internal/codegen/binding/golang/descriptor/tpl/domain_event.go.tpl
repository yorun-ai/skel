{{- define "domainEvents" -}}
{{- if .Descriptor.Events }}
	Events: []*descriptor.Event{
		{{- range $event := .Descriptor.Events }}
		{{ template "eventDescriptorValue" $event }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "eventDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Ext }}, Ext: true{{ end }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ template "memberDescriptorList" .Members }}}
{{- end }}
