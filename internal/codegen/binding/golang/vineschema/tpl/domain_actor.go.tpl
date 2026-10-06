{{- define "domainActors" -}}
{{- if .Schema.Actors }}
	Actors: []*skel.ActorSchema{
		{{- range $actor := .Schema.Actors }}
		{{ template "actorSchemaValue" $actor }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "actorSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Vias }}, Vias: []skel.ActorVia{ {{- range $via := .Vias }}{{ viaLiteral $via }}, {{- end }} }{{ end }}, AuthEnabled: {{ .AuthEnabled }}{{ if .IdentifierField }}, IdentifierField: {{ quote .IdentifierField }}{{ end }}{{ if .AuthCredential }}, AuthCredential: {{ template "dataSchema" .AuthCredential }}{{ end }}{{ if .AuthInfo }}, AuthInfo: {{ template "dataSchema" .AuthInfo }}{{ end }}{{ if .AuthService }}, AuthService: {{ template "serviceSchema" .AuthService }}{{ end }}{{ if .AuthMethod }}, AuthMethod: {{ template "methodSchema" .AuthMethod }}{{ end }}, PermEnabled: {{ .PermEnabled }}{{ if .PermService }}, PermService: {{ template "serviceSchema" .PermService }}{{ end }}{{ if .PermMethod }}, PermMethod: {{ template "methodSchema" .PermMethod }}{{ end }}}
{{- end }}
