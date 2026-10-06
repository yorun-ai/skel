{{- define "domainResources" -}}
{{- if .Descriptor.Resources }}
	Resources: []*descriptor.Resource{
		{{- range $resource := .Descriptor.Resources }}
		{{ template "resourceDescriptorValue" $resource }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "resourceDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .CheckService }}, CheckService: {{ template "serviceDescriptor" .CheckService }}{{ end }}{{ template "resourceCheckDescriptorList" .Checks }}{{ template "resourceActionDescriptorList" .Actions }}}
{{- end }}

{{- define "resourceCheckDescriptorList" -}}
{{- if . }}, Checks: []*descriptor.ResourceCheck{
{{- range $check := . }}
{{ template "resourceCheckDescriptorValue" $check }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "resourceCheckDescriptorValue" -}}
{Name: {{ quote .Name }}{{ template "deprecatedFields" . }}, Method: {{ template "methodDescriptor" .Method }}{{ template "argumentDescriptorList" .Arguments }}}
{{- end }}

{{- define "resourceActionDescriptorList" -}}
{{- if . }}, Actions: []*descriptor.ResourceAction{
{{- range $action := . }}
{{ template "resourceActionDescriptorValue" $action }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "resourceActionDescriptorValue" -}}
{Name: {{ quote .Name }}, PermissionCode: {{ quote .PermissionCode }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}{{ template "resourceCheckDescriptorList" .Checks }}}
{{- end }}
