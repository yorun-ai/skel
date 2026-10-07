{{- define "serviceDescriptor" -}}
&descriptor.Service{{ template "serviceDescriptorValue" . }}
{{- end }}

{{- define "serviceDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Api }}, Api: true{{ end }}{{ if .Ext }}, Ext: true{{ end }}, AuthMode: {{ authLiteral .AuthMode }}{{ if .Audiences }}, Audiences: []*descriptor.ActorAudience{ {{- range $actor := .Audiences }}{{ template "actorAudienceDescriptor" $actor }}, {{- end }} }{{ end }}{{ if .Require }}, Require: {{ template "permissionRequire" .Require }}{{ end }}{{ template "methodDescriptorList" .Methods }}}
{{- end }}

{{- define "actorAudienceDescriptor" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ with .Via }}, Via: {{ viaLiteral . }}{{ end }}}
{{- end }}

{{- define "methodDescriptorList" -}}
{{- if . }}, Methods: []*descriptor.Method{
{{- range $method := . }}
{{ template "methodDescriptorValue" $method }},
{{- end }}
}{{ end -}}
{{- end }}
