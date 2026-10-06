{{- define "domainEnums" -}}
{{- if .Schema.Enums }}
	Enums: []*skel.EnumSchema{
		{{- range $enum := .Schema.Enums }}
		{{ template "enumSchemaValue" $enum }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "enumSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Items }}, Items: []*skel.EnumItemSchema{
{{- range $item := .Items }}
{{ template "enumItemSchemaValue" $item }},
{{- end }}
}{{ end }}}
{{- end }}

{{- define "enumItemSchemaValue" -}}
{Name: {{ quote .Name }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}}
{{- end }}
