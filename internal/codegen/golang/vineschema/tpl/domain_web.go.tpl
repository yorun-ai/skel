{{- define "domainWebs" -}}
{{- if .Schema.Webs }}
	Webs: []*skel.WebSchema{
		{{- range $web := .Schema.Webs }}
		{{ template "webSchemaValue" $web }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "webSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .AuthMode }}, AuthMode: {{ authLiteral .AuthMode }}{{ end }}{{ if .MountPath }}, MountPath: {{ quote .MountPath }}{{ end }}{{ if .Audiences }}, Audiences: []*skel.ActorAudienceSchema{ {{- range $actor := .Audiences }}{{ template "actorAudienceSchema" $actor }}, {{- end }} }{{ end }}}
{{- end }}
