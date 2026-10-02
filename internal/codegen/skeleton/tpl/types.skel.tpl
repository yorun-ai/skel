domain {{ .Domain.Name }}
{{ if .Imports }}
{{ range $import := .Imports -}}
import {{ $import.Name }}{{ with importAlias $import }} as {{ . }}{{ end }}
{{ end -}}
{{ end }}
{{ range $i, $enum := .Enums -}}
{{ if $i }}
{{ end -}}
{{ template "description" (description $enum.Description 0) }}{{ template "deprecated" (deprecated $enum.Deprecated $enum.DeprecatedReason 0) }}{{ if $enum.Pub }}pub {{ end }}enum {{ $enum.Name }} {
{{- range $item := $enum.Items }}
{{ template "description" (description $item.Description 4) }}{{ template "deprecated" (deprecated $item.Deprecated $item.DeprecatedReason 4) }}    {{ $item.Name }}
{{- end }}
}
{{ end -}}
{{ range $i, $data := .Data -}}
{{ if or $.Enums $i }}
{{ end -}}
{{ template "description" (description $data.Description 0) }}{{ template "deprecated" (deprecated $data.Deprecated $data.DeprecatedReason 0) }}{{ template "sensitive" (sensitive $data.Sensitive 0) }}{{ if $data.Pub }}pub {{ end }}data {{ $data.Name }}{{ with typeParams $data.TypeParameters }}<{{ range $i, $name := . }}{{ if $i }}, {{ end }}{{ $name }}{{ end }}>{{ end }} {
{{- range $member := $data.Members }}
{{ template "description" (description $member.Description 4) }}{{ template "deprecated" (deprecated $member.Deprecated $member.DeprecatedReason 4) }}{{ template "example" (example $member.Example 4) }}{{ template "sensitive" (sensitive $member.Sensitive 4) }}    {{ $member.Name }}: {{ template "type" (typeRef $member.Type) }}
{{- end }}
}
{{ end -}}
{{ range $i, $config := .Configs -}}
{{ if or $.Enums $.Data $i }}
{{ end -}}
{{ template "description" (description $config.Description 0) }}{{ template "deprecated" (deprecated $config.Deprecated $config.DeprecatedReason 0) }}{{ template "sensitive" (sensitive $config.Sensitive 0) }}pub config {{ $config.Name }}{{ with configSuffix $config }} {{ . }}{{ end }} {
{{- range $member := $config.Members }}
{{ template "description" (description $member.Description 4) }}{{ template "deprecated" (deprecated $member.Deprecated $member.DeprecatedReason 4) }}{{ template "example" (example $member.Example 4) }}{{ template "sensitive" (sensitive $member.Sensitive 4) }}    {{ $member.Name }}: {{ template "type" (typeRef $member.Type) }}
{{- end }}
}
{{ end -}}
{{ range $i, $resource := .Resources -}}
{{ if or $.Enums $.Data $.Configs $i }}
{{ end -}}
{{ template "description" (description $resource.Description 0) }}{{ template "deprecated" (deprecated $resource.Deprecated $resource.DeprecatedReason 0) }}{{ if $resource.Pub }}pub {{ end }}resource {{ $resource.Name }} {
{{ range $check := $resource.Checks }}{{ template "resourceCheck" (resourceCheck $check 4) }}
{{ end }}
{{ if and $resource.Checks $resource.Actions }}
{{ end -}}
{{ range $i, $action := $resource.Actions }}{{ if $i }}
{{ end }}
{{ template "description" (description $action.Description 4) }}{{ template "deprecated" (deprecated $action.Deprecated $action.DeprecatedReason 4) }}    action {{ $action.Name }}{{ if $action.Checks }} {
{{ range $check := $action.Checks }}{{ template "resourceCheck" (resourceCheck $check 8) }}
{{- end }}
    }{{ end }}
{{- end }}
}
{{ end -}}
