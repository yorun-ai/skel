{{- define "dataSchema" -}}
&skel.DataSchema{{ template "dataSchemaValue" . }}
{{- end }}

{{- define "dataSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ if .TypeParameters }}, TypeParameters: []string{ {{- range $typeParameter := .TypeParameters }}{{ quote $typeParameter }}, {{- end }} }{{ end }}{{ template "memberSchemaList" .Members }}}
{{- end }}
