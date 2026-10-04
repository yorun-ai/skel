domain {{ .Domain.Name }}
{{ if .Imports }}
{{ range $import := .Imports -}}
import {{ $import.Name }}{{ with importAlias $import }} as {{ . }}{{ end }}
{{ end -}}
{{ end }}
{{ range $i, $event := .Events -}}
{{ if $i }}
{{ end -}}
{{ template "description" (description $event.Description 0) }}{{ template "deprecated" (deprecated $event.Deprecated $event.DeprecatedReason 0) }}{{ if $event.Ext }}ext{{ else }}pub{{ end }} event {{ $event.Name }} {
{{ template "sensitive" (sensitive $event.Sensitive 4) }}    payload {
{{- range $member := $event.Members }}
{{ template "description" (description $member.Description 8) }}{{ template "deprecated" (deprecated $member.Deprecated $member.DeprecatedReason 8) }}{{ template "example" (example $member.Example 8) }}{{ template "sensitive" (sensitive $member.Sensitive 8) }}        {{ $member.Name }}: {{ template "type" (typeRef $member.Type) }}
{{- end }}
    }
}
{{ end -}}
