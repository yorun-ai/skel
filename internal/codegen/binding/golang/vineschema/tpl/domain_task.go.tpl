{{- define "domainTasks" -}}
{{- if .Schema.Tasks }}
	Tasks: []*skel.TaskSchema{
		{{- range $task := .Schema.Tasks }}
		{{ template "taskSchemaValue" $task }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "taskSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ template "triggerSchemaList" .Triggers }}}
{{- end }}

{{- define "triggerSchemaList" -}}
{{- if . }}, Triggers: []*skel.TriggerSchema{
{{- range $trigger := . }}
{{ template "triggerSchemaValue" $trigger }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "triggerSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .InputDescription }}, InputDescription: {{ quote .InputDescription }}{{ end }}{{ if .ArgumentsSensitive }}, ArgumentsSensitive: true{{ end }}{{ template "argumentSchemaList" .Arguments }}}
{{- end }}
