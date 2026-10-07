{{- range $import := $.TypeImports }}
import type * as {{ $import.Alias }} from '{{ $import.Path }}';
{{- end }}
{{- if $.TypeImports }}

{{- end }}
{{- range $index, $e := $.Enums }}
{{- if gt $index 0 }}

{{- end }}
{{- if $e.CommentLines }}
{{- if eq (len $e.CommentLines) 1 }}
/** {{ index $e.CommentLines 0 }} */
{{- else }}
/**
{{- range $line := $e.CommentLines }}
 * {{ $line }}
{{- end }}
 */
{{- end }}
{{- end }}
export type {{ $e.Name }} =
{{- range $item := $e.Items }}
  {{- if $item.CommentLines }}
  {{- if eq (len $item.CommentLines) 1 }}
  /** {{ index $item.CommentLines 0 }} */
  {{- else }}
  /**
  {{- range $line := $item.CommentLines }}
   * {{ $line }}
  {{- end }}
   */
  {{- end }}
  {{- end }}
  | {{ $item.Literal }}{{ $item.ValuePadding }}
{{- end }}
;
{{- end }}
{{- if and $.Enums $.Data }}

{{- end }}
{{- range $index, $s := $.Data }}
{{- if gt $index 0 }}

{{- end }}
{{- if $s.CommentLines }}
{{- if eq (len $s.CommentLines) 1 }}
/** {{ index $s.CommentLines 0 }} */
{{- else }}
/**
{{- range $line := $s.CommentLines }}
 * {{ $line }}
{{- end }}
 */
{{- end }}
{{- end }}
{{ if not $s.Members -}}
export type {{ $s.FullName }} = {}
{{- else -}}
export type {{ $s.FullName }} = {{ "{" }}{{ range $sm := $s.Members }}
  {{- if $sm.CommentLines }}
  {{- if eq (len $sm.CommentLines) 1 }}
  /** {{ index $sm.CommentLines 0 }} */
  {{- else }}
  /**
  {{- range $line := $sm.CommentLines }}
   * {{ $line }}
  {{- end }}
   */
  {{- end }}
  {{- end }}
  {{ $sm.Name }}:{{ $sm.NamePadding }} {{ $sm.Type.Plain }};{{ end }}
}
{{- end }}
{{- end }}
{{- if or $.Enums $.Data }}

{{- end }}
export {};
