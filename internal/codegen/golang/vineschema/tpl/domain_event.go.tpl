{{- define "domainEvents" -}}
{{- if .Schema.Events }}
	Events: []*skel.EventSchema{
		{{- range $event := .Schema.Events }}
		{{ template "eventSchemaValue" $event }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "eventSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Ext }}, Ext: true{{ end }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ template "memberSchemaList" .Members }}}
{{- end }}
