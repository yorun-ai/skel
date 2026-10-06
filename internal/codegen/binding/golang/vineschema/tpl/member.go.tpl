{{- define "memberSchema" -}}
{Name: {{ quote .Name }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}{{ if .Example }}, Example: {{ quote .Example }}{{ end }}{{ if .Sensitive }}, Sensitive: true{{ end }}, Type: {{ template "typeSchema" .Type }}}
{{- end }}

{{- define "memberSchemaList" -}}
{{- if . }}, Members: []*skel.MemberSchema{
{{- range $member := . }}
{{ template "memberSchema" $member }},
{{- end }}
}{{ end -}}
{{- end }}
