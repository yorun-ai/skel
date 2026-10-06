package {{ $.PackageName }}{{ template "imports" . }}

func init() { {{ range $service := $.Services }}
	rpc.Register({{ $service.SpecName }}){{ end }}
}
{{ range $service := $.Services }}
{{- if $service.CommentLines }}
{{- range $line := $service.CommentLines }}
// {{ $line }}
{{- end }}
{{- else }}
// {{ $service.ServerName }}
{{- end }}

{{ template "serviceInfo" $service -}}
{{ template "serviceArguments" $service -}}
{{ template "serviceServer" $service -}}
{{ template "serviceERServer" $service -}}
{{ template "serviceClient" $service -}}
{{ template "serviceERClient" $service }}
{{- end }}
