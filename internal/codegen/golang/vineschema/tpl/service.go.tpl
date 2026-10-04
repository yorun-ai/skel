{{- define "serviceSchema" -}}
&skel.ServiceSchema{{ template "serviceSchemaValue" . }}
{{- end }}

{{- define "serviceSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}, Pub: {{ .Pub }}{{ if .Api }}, Api: true{{ end }}{{ if .Ext }}, Ext: true{{ end }}, AuthMode: {{ authLiteral .AuthMode }}{{ if .Audiences }}, Audiences: []*skel.ActorAudienceSchema{ {{- range $actor := .Audiences }}{{ template "actorAudienceSchema" $actor }}, {{- end }} }{{ end }}{{ if .Require }}, Require: {{ template "permissionRequire" .Require }}{{ end }}{{ template "methodSchemaList" .Methods }}}
{{- end }}

{{- define "actorAudienceSchema" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ with .Via }}, Via: {{ viaLiteral . }}{{ end }}}
{{- end }}

{{- define "methodSchemaList" -}}
{{- if . }}, Methods: []*skel.MethodSchema{
{{- range $method := . }}
{{ template "methodSchemaValue" $method }},
{{- end }}
}{{ end -}}
{{- end }}
