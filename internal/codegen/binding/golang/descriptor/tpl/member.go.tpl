{{- define "memberDescriptor" -}}
{Name: {{ quote .Name }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}{{ if .Example }}, Example: {{ quote .Example }}{{ end }}{{ if .Sensitive }}, Sensitive: true{{ end }}, Type: {{ template "typeDescriptor" .Type }}}
{{- end }}

{{- define "memberDescriptorList" -}}
{{- if . }}, Members: []*descriptor.Member{
{{- range $member := . }}
{{ template "memberDescriptor" $member }},
{{- end }}
}{{ end -}}
{{- end }}
