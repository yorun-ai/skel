{{- define "dataDescriptor" -}}
&descriptor.Data{{ template "dataDescriptorValue" . }}
{{- end }}

{{- define "dataDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Sensitive }}, Sensitive: true{{ end }}{{ if .TypeParameters }}, TypeParameters: []string{ {{- range $typeParameter := .TypeParameters }}{{ quote $typeParameter }}, {{- end }} }{{ end }}{{ template "memberDescriptorList" .Members }}}
{{- end }}
