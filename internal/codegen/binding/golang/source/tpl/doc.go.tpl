{{- if .CommentLines }}{{- range $line := .CommentLines }}{{ printf "// %s\n" $line }}{{- end }}{{- else }}{{ printf "// %s\n" .PackageName }}{{- end -}}
//
// This package is fully managed by skelc. Do not edit generated files or add
// handwritten Go files to it. Neither skelc nor Vine guarantees compilation,
// runtime behavior, or compatibility for generated packages containing
// unmanaged files.
package {{ .PackageName }}
