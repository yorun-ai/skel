{{ define "apiInfo" }}{{ $s := . }}
var (
    {{ .SpecName }} = &vrpc.ServiceSpec{
        Name: "{{ .Name }}",
        SkelName: "{{ .SkelName }}",
        Methods: []vrpc.MethodSpec{
            {{ range .Methods }}{Name: "{{ .Name }}", SkelName: "{{ .SkelName }}", ArgumentsContainsBinaryType: {{ .ArgumentsBinary }}, ResultContainsBinaryType: {{ .ResultBinary }}, ArgumentsSensitive: {{ .ArgumentsSensitive }}, ResultSensitive: {{ .ResultSensitive }}},
            {{ end -}}
        },
    }
    {{ range .Methods }}_{{ $s.Name }}{{ .Name }}Method vrpc.MethodInfo
    {{ end }}
)
{{ end }}
