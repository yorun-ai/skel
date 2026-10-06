{{- define "domainResources" -}}
{{- if .Schema.Resources }}
	Resources: []*skel.ResourceSchema{
		{{- range $resource := .Schema.Resources }}
		{{ template "resourceSchemaValue" $resource }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "resourceSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .CheckService }}, CheckService: {{ template "serviceSchema" .CheckService }}{{ end }}{{ template "resourceCheckSchemaList" .Checks }}{{ template "resourceActionSchemaList" .Actions }}}
{{- end }}

{{- define "resourceCheckSchemaList" -}}
{{- if . }}, Checks: []*skel.ResourceCheckSchema{
{{- range $check := . }}
{{ template "resourceCheckSchemaValue" $check }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "resourceCheckSchemaValue" -}}
{Name: {{ quote .Name }}{{ template "deprecatedFields" . }}, Method: {{ template "methodSchema" .Method }}{{ template "argumentSchemaList" .Arguments }}}
{{- end }}

{{- define "resourceActionSchemaList" -}}
{{- if . }}, Actions: []*skel.ResourceActionSchema{
{{- range $action := . }}
{{ template "resourceActionSchemaValue" $action }},
{{- end }}
}{{ end -}}
{{- end }}

{{- define "resourceActionSchemaValue" -}}
{Name: {{ quote .Name }}, PermissionCode: {{ quote .PermissionCode }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}{{ template "resourceCheckSchemaList" .Checks }}}
{{- end }}
