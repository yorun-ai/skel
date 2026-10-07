{{- define "methodDescriptor" -}}
&descriptor.Method{{ template "methodDescriptorValue" . }}
{{- end }}

{{- define "methodDescriptorValue" -}}
{Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ if .Description }}, Description: {{ quote .Description }}{{ end }}{{ template "deprecatedFields" . }}, Hash: {{ quote .Hash }}{{ if .Example }}, Example: {{ quote .Example }}{{ end }}, AuthMode: {{ authLiteral .AuthMode }}{{ if .Require }}, Require: {{ template "permissionRequire" .Require }}{{ end }}, EffectiveAuthMode: {{ authLiteral .EffectiveAuthMode }}{{ if .EffectiveRequire }}, EffectiveRequire: {{ template "permissionRequire" .EffectiveRequire }}{{ end }}{{ if .InputDescription }}, InputDescription: {{ quote .InputDescription }}{{ end }}{{ if .ArgumentsSensitive }}, ArgumentsSensitive: true{{ end }}{{ if .OutputDescription }}, OutputDescription: {{ quote .OutputDescription }}{{ end }}{{ if .OutputExample }}, OutputExample: {{ quote .OutputExample }}{{ end }}{{ if .ResultSensitive }}, ResultSensitive: true{{ end }}{{ if .ResultType }}, ResultType: {{ template "typeDescriptor" .ResultType }}{{ end }}{{ template "argumentDescriptorList" .Arguments }}}
{{- end }}

{{- define "argumentDescriptorList" -}}
{{- if . }}, Arguments: []*descriptor.Member{
{{- range $argument := . }}
{{ template "memberDescriptor" $argument }},
{{- end }}
}{{ end -}}
{{- end }}
