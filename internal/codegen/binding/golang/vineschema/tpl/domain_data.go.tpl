{{- define "domainData" -}}
{{- if .Schema.Data }}
	Data: []*skel.DataSchema{
		{{- range $data := .Schema.Data }}
		{{ template "dataSchemaValue" $data }},
		{{- end }}
	},
	{{- end }}
{{- end }}
