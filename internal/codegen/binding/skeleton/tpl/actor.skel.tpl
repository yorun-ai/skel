domain {{ .Domain.Name }}
{{ if .Imports }}
{{ range $import := .Imports -}}
import {{ $import.Name }}{{ with importAlias $import }} as {{ . }}{{ end }}
{{ end -}}
{{ end }}
{{ range $i, $actor := .Actors -}}
{{ if $i }}
{{ end -}}
{{ template "description" (description $actor.Description 0) }}{{ template "deprecated" (deprecated $actor.Deprecated $actor.DeprecatedReason 0) }}pub actor {{ $actor.Name }} {
{{- range $via := $actor.Vias }}
    via {{ $via.Name }} {}
{{- end }}
{{- if $actor.AuthEnabled }}
    auth {
{{- with $actor.AuthCredential }}
{{ template "sensitive" (sensitive .Sensitive 8) }}        credential {
{{- range $member := .Members }}
{{ template "description" (description $member.Description 12) }}{{ template "deprecated" (deprecated $member.Deprecated $member.DeprecatedReason 12) }}{{ template "example" (example $member.Example 12) }}{{ template "sensitive" (sensitive $member.Sensitive 12) }}            {{ $member.Name }}: {{ template "type" (typeRef $member.Type) }}
{{- end }}
        }
{{- end }}
{{- with $actor.AuthInfo }}
{{ template "sensitive" (sensitive .Sensitive 8) }}        info {
{{- range $member := .Members }}
{{ template "description" (description $member.Description 12) }}{{ template "deprecated" (deprecated $member.Deprecated $member.DeprecatedReason 12) }}{{ template "example" (example $member.Example 12) }}{{ template "sensitive" (sensitive $member.Sensitive 12) }}{{ if eq $member.Name $actor.IdentifierField }}            @identifier
{{ end }}            {{ $member.Name }}: {{ template "type" (typeRef $member.Type) }}
{{- end }}
        }
{{- end }}
    }
{{- end }}
{{- if $actor.PermEnabled }}
    permission {}
{{- end }}
}
{{ end -}}
