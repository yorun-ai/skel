{{- define "domainData" -}}
{{- if .Descriptor.Data }}
	Data: []*descriptor.Data{
		{{- range $data := .Descriptor.Data }}
		{{ template "dataDescriptorValue" $data }},
		{{- end }}
	},
	{{- end }}
{{- end }}
