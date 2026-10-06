{{ define "description" }}{{ with . }}{{ $indent := spaces " " .Indent }}{{ if .Multiline }}{{ $indent }}@desc("""
{{ range .Lines }}{{ $indent }}{{ . }}
{{ end }}{{ $indent }}""")
{{ else }}{{ $indent }}@desc({{ .Quoted }})
{{ end }}{{ end }}{{ end }}
{{ define "example" }}{{ with . }}{{ $indent := spaces " " .Indent }}{{ if .Multiline }}{{ if .Object }}{{ $indent }}@example({
{{ range .Lines }}{{ $indent }}    {{ . }}
{{ end }}{{ $indent }}})
{{ else }}{{ $indent }}@example(
{{ range .Lines }}{{ $indent }}{{ . }}
{{ end }}{{ $indent }})
{{ end }}{{ else }}{{ $indent }}@example({{ .Quoted }})
{{ end }}{{ end }}{{ end }}
{{ define "sensitive" }}{{ with . }}{{ spaces " " .Indent }}@sensitive
{{ end }}{{ end }}
{{ define "deprecated" }}{{ with . }}{{ $indent := spaces " " .Indent }}{{ if .Multiline }}{{ $indent }}@deprecated("""
{{ range .Lines }}{{ $indent }}{{ . }}
{{ end }}{{ $indent }}""")
{{ else }}{{ $indent }}@deprecated({{ .Quoted }})
{{ end }}{{ end }}{{ end }}
{{ define "resourceCheck" }}{{ template "description" (description .Description .Indent) }}{{ template "deprecated" (deprecated .Deprecated .DeprecatedReason .Indent) }}{{ spaces " " .Indent }}check {{ .Name }} {{ if or .Arguments .InputDescription .InputSensitive }}{
{{ template "description" (description .InputDescription .InputIndent) }}{{ template "sensitive" (sensitive .InputSensitive .InputIndent) }}{{ spaces " " .InputIndent }}input {
{{ range $arg := .Arguments }}{{ template "description" (description $arg.Description $.ArgumentIndent) }}{{ template "deprecated" (deprecated $arg.Deprecated $arg.DeprecatedReason $.ArgumentIndent) }}{{ template "example" (example $arg.Example $.ArgumentIndent) }}{{ template "sensitive" (sensitive $arg.Sensitive $.ArgumentIndent) }}{{ spaces " " $.ArgumentIndent }}{{ $arg.Name }}: {{ template "type" (typeRef $arg.Type) }}
{{ end }}{{ spaces " " .InputIndent }}}
{{ spaces " " .Indent }}}{{ else }}{}{{ end }}
{{ end }}
{{ define "type" }}{{ if eq .Kind "list" }}list<{{ template "type" .Value }}>{{ else if eq .Kind "map" }}map<{{ template "type" .Key }}, {{ template "type" .Value }}>{{ else }}{{ with .Qualifier }}{{ . }}.{{ end }}{{ .Name }}{{ with .Arguments }}<{{ range $i, $argument := . }}{{ if $i }}, {{ end }}{{ template "type" $argument }}{{ end }}>{{ end }}{{ end }}{{ if .Nullable }}?{{ end }}{{ end }}
