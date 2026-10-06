{{- define "typeSchema" -}}
{{- if . -}}
&skel.TypeSchema{ {{- template "typeSchemaFields" . }} }
{{- else -}}
nil
{{- end -}}
{{- end }}

{{- define "typeSchemaValue" -}}
{ {{- template "typeSchemaFields" . }} }
{{- end }}

{{- define "typeArguments" -}}
{{- if . }}, TypeArguments: []*skel.TypeSchema{ {{- range $typeArgument := . }}{{ template "typeSchemaValue" $typeArgument }}, {{- end }} }{{ end -}}
{{- end }}

{{- define "typeSchemaFields" -}}
{{- if eq .Kind "scalar" -}}
Kind: skel.TypeKindScalar, Scalar: {{ scalarLiteral .Scalar }}
{{- else if eq .Kind "list" -}}
Kind: skel.TypeKindList, Element: {{ template "typeSchema" .Element }}
{{- else if eq .Kind "map" -}}
Kind: skel.TypeKindMap, Key: {{ template "typeSchema" .Key }}, Value: {{ template "typeSchema" .Value }}
{{- else if eq .Kind "enum" -}}
Kind: skel.TypeKindEnum, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}
{{- else if eq .Kind "data" -}}
Kind: skel.TypeKindData, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "config" -}}
Kind: skel.TypeKindConfig, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "event" -}}
Kind: skel.TypeKindEvent, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "typeParameter" -}}
Kind: skel.TypeKindTypeParameter, Name: {{ quote .Name }}
{{- end -}}
{{- if .Nullable }}, Nullable: true{{ end -}}
{{- end }}
