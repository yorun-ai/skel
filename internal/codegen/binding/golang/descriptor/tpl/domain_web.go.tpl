{{- define "domainWebs" -}}
{{- if .Descriptor.Webs }}
	Webs: []*descriptor.Web{
		{{- range $web := .Descriptor.Webs }}
		{{ template "webDescriptorValue" $web }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "webDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .AuthMode }}, AuthMode: {{ authLiteral .AuthMode }}{{ end }}{{ if .MountPath }}, MountPath: {{ quote .MountPath }}{{ end }}{{ if .Audiences }}, Audiences: []*descriptor.ActorAudience{ {{- range $actor := .Audiences }}{{ template "actorAudienceDescriptor" $actor }}, {{- end }} }{{ end }}}
{{- end }}
