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
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Vias }}, Vias: []descriptor.ActorVia{ {{- range $via := .Vias }}{{ viaLiteral $via }}, {{- end }} }{{ end }}{{ with .Auth }}, Auth: &descriptor.ActorAuth{Credential: {{ template "dataDescriptor" .Credential }}, Info: {{ template "dataDescriptor" .Info }}{{ if .IdentifierField }}, IdentifierField: {{ quote .IdentifierField }}{{ end }}, Service: {{ template "serviceDescriptor" .Service }}, MethodName: {{ quote .MethodName }} }{{ end }}{{ with .Permission }}, Permission: &descriptor.ActorPermission{Service: {{ template "serviceDescriptor" .Service }}, MethodName: {{ quote .MethodName }} }{{ end }}}
{{- end }}
