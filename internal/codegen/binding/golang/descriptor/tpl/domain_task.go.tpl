{{- define "domainTasks" -}}
{{- if .Descriptor.Tasks }}
	Tasks: []*descriptor.Task{
		{{- range $task := .Descriptor.Tasks }}
		{{ template "taskDescriptorValue" $task }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "taskDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ template "triggerDescriptorList" .Triggers }}}
{{- end }}

{{- define "triggerDescriptorList" -}}
{{- if . }}, Triggers: []*descriptor.TaskTrigger{
{{- range $trigger := . }}
{{ template "triggerDescriptorValue" $trigger }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "triggerDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .InputDescription }}, InputDescription: {{ quote .InputDescription }}{{ end }}{{ if .ArgumentsSensitive }}, ArgumentsSensitive: true{{ end }}{{ template "argumentDescriptorList" .Arguments }}}
{{- end }}
