package {{ $.PackageName }}{{ template "imports" . }}

{{ if $.Services -}}
func init() { {{ range $service := $.Services }}
	rpc.Register({{ $service.SpecName }}){{ end }}
}

{{ end -}}
{{ range $resource := $.Resources -}}
{{ if $resource.Actions -}}
const (
{{- range $action := $resource.Actions }}
	{{- if $action.CommentLines }}
	{{- range $line := $action.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $action.PermissionName }} string = "{{ $action.PermissionCode }}"
{{- end }}
)

{{ end -}}
{{ end -}}
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
