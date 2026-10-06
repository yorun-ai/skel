{{- define "domainEnums" -}}
{{- if .Descriptor.Enums }}
	Enums: []*descriptor.Enum{
		{{- range $enum := .Descriptor.Enums }}
		{{ template "enumDescriptorValue" $enum }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "enumDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Items }}, Items: []*descriptor.EnumItem{
{{- range $item := .Items }}
{{ template "enumItemDescriptorValue" $item }},
{{- end }}
}{{ end }}}
{{- end }}

{{- define "enumItemDescriptorValue" -}}
{Name: {{ quote .Name }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}}
{{- end }}
