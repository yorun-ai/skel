{{- define "methodSchema" -}}
&skel.MethodSchema{{ template "methodSchemaValue" . }}
{{- end }}

{{- define "methodSchemaValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Example }}, Example: {{ quote .Example }}{{ end }}, AuthMode: {{ authLiteral .AuthMode }}{{ if .Require }}, Require: {{ template "permissionRequire" .Require }}{{ end }}{{ if .InputDescription }}, InputDescription: {{ quote .InputDescription }}{{ end }}{{ if .ArgumentsSensitive }}, ArgumentsSensitive: true{{ end }}{{ if .OutputDescription }}, OutputDescription: {{ quote .OutputDescription }}{{ end }}{{ if .OutputExample }}, OutputExample: {{ quote .OutputExample }}{{ end }}{{ if .ResultSensitive }}, ResultSensitive: true{{ end }}{{ if .ResultType }}, ResultType: {{ template "typeSchema" .ResultType }}{{ end }}{{ template "argumentSchemaList" .Arguments }}}
{{- end }}

{{- define "argumentSchemaList" -}}
{{- if . }}, Arguments: []*skel.MemberSchema{
{{- range $argument := . }}
{{ template "memberSchema" $argument }},
{{- end }}
}{{ end -}}
{{- end }}
