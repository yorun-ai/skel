{{- define "deprecatedFields" -}}
{{- if .Deprecated }}, Deprecated: true, DeprecatedReason: {{ quote .DeprecatedReason }}{{ end -}}
{{- end }}
