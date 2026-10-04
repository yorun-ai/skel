package {{ .PackageName }}
{{ template "imports" . }}
func init() {
    {{ range $s := .Services }}vrpc.Register({{ .SpecName }})
    {{ range .Methods }}_{{ $s.Name }}{{ .Name }}Method = vrpc.MustGetMethodInfo("{{ $s.SkelName }}", "{{ .SkelName }}")
    {{ end }}{{ end }}
}
{{ range .Services }}
{{ template "apiInfo" . }}
{{ template "serviceArguments" . }}
{{ template "apiClient" . }}
{{ end }}
