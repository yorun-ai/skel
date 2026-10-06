{{- define "domainActors" -}}
{{- if .Descriptor.Actors }}
	Actors: []*descriptor.Actor{
		{{- range $actor := .Descriptor.Actors }}
		{{ template "actorDescriptorValue" $actor }},
		{{- end }}
	},
	{{- end }}
{{- end }}

{{- define "actorDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Vias }}, Vias: []descriptor.ActorVia{ {{- range $via := .Vias }}{{ viaLiteral $via }}, {{- end }} }{{ end }}, AuthEnabled: {{ .AuthEnabled }}{{ if .IdentifierField }}, IdentifierField: {{ quote .IdentifierField }}{{ end }}{{ if .AuthCredential }}, AuthCredential: {{ template "dataDescriptor" .AuthCredential }}{{ end }}{{ if .AuthInfo }}, AuthInfo: {{ template "dataDescriptor" .AuthInfo }}{{ end }}{{ if .AuthService }}, AuthService: {{ template "serviceDescriptor" .AuthService }}{{ end }}{{ if .AuthMethod }}, AuthMethod: {{ template "methodDescriptor" .AuthMethod }}{{ end }}, PermissionEnabled: {{ .PermissionEnabled }}{{ if .PermissionService }}, PermissionService: {{ template "serviceDescriptor" .PermissionService }}{{ end }}{{ if .PermissionMethod }}, PermissionMethod: {{ template "methodDescriptor" .PermissionMethod }}{{ end }}}
{{- end }}
