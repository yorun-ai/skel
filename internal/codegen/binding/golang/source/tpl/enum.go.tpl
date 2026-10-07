package {{ $.PackageName }}{{ template "imports" . }}
{{ range $e := $.Enums }}
{{- if $e.CommentLines }}
{{- range $line := $e.CommentLines }}
// {{ $line }}
{{- end }}
{{- end }}
type {{ $e.Name }} string

const (
	{{ $e.UnspecifiedItem.Name }} {{ $e.Name }} = "{{ $e.UnspecifiedItem.Value }}"
{{- range $ei := $e.Items }}
	{{- if $ei.CommentLines }}
	{{- range $line := $ei.CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
	{{ $ei.Name }} {{ $e.Name }} = "{{ $ei.Value }}"
{{- end }}
)

func ({{ $e.VarName }} {{ $e.Name }}) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "\"%s\"", {{ $e.VarName }}), nil
}

func ({{ $e.VarName }} *{{ $e.Name }}) UnmarshalJSON(data []byte) error {
	var err error = nil
	switch string(data) {
{{- range $ei := $e.Items }}
	case `"{{ $ei.Value }}"`:
		*{{ $e.VarName }} = {{ $ei.Name }}
{{- end }}
	case "null":
		err = fmt.Errorf("unexpected null value for non-pointer {{ $e.Name }}")
	default:
		*{{ $e.VarName }} = {{ $e.UnspecifiedItem.Name }}
	}
	return err
}
{{ end }}
