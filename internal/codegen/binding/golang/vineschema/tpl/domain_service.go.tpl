{{- define "domainServices" -}}
{{- if .Schema.Services }}
	Services: []*skel.ServiceSchema{
		{{- range $service := .Schema.Services }}
		{{ template "serviceSchemaValue" $service }},
		{{- end }}
	},
	{{- end }}
{{- end }}
