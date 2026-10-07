{{ define "importSpec" }}{{ with .Alias }}{{ . }} {{ end }}{{ printf "%q" .Path }}{{ end }}
{{ define "imports" }}{{ if or .StdImports .ModuleImports }}

{{ if and (eq (len .StdImports) 1) (not .ModuleImports) }}import {{ template "importSpec" (index .StdImports 0) }}
{{ else if and (not .StdImports) (eq (len .ModuleImports) 1) }}import {{ template "importSpec" (index .ModuleImports 0) }}
{{ else }}import (
{{ range .StdImports }}	{{ template "importSpec" . }}
{{ end }}{{ if and .StdImports .ModuleImports }}
{{ end }}{{ range .ModuleImports }}	{{ template "importSpec" . }}
{{ end }})
{{ end }}{{ end }}{{ end }}
