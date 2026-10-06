{{- define "serviceArguments" -}}{{ if .HasMethodArguments }}
// {{ .Name }} / Arguments
{{ range .Methods }}{{ if .ArgumentsData }}
type {{ .ArgumentsData.Name }} struct {
{{- range $index, $argument := .ArgumentsData.Members }}
	{{ $argument.Name }} {{ $argument.Type.Plain }} `json:"{{ $argument.SkelName }}" skel:"index({{ $index }}){{ if $argument.Sensitive }},sensitive{{ end }}"`
{{- end }}
}
{{ end }}{{- end -}}
{{ end }}{{ end }}
