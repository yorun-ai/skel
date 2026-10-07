{{- define "domainServices" -}}
{{- if .Descriptor.Services }}
	Services: []*descriptor.Service{
		{{- range $service := .Descriptor.Services }}
		{{ template "serviceDescriptorValue" $service }},
		{{- end }}
	},
	{{- end }}
{{- end }}
