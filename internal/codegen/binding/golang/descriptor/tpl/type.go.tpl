{{- define "typeDescriptor" -}}
{{- if . -}}
&descriptor.Type{ {{- template "typeDescriptorFields" . }} }
{{- else -}}
nil
{{- end -}}
{{- end }}

{{- define "typeDescriptorValue" -}}
{ {{- template "typeDescriptorFields" . }} }
{{- end }}

{{- define "typeArguments" -}}
{{- if . }}, TypeArguments: []*descriptor.Type{ {{- range $typeArgument := . }}{{ template "typeDescriptorValue" $typeArgument }}, {{- end }} }{{ end -}}
{{- end }}

{{- define "typeDescriptorFields" -}}
{{- if eq .Kind "scalar" -}}
Kind: descriptor.TypeKindScalar, Scalar: {{ scalarLiteral .Scalar }}
{{- else if eq .Kind "list" -}}
Kind: descriptor.TypeKindList, Element: {{ template "typeDescriptor" .Element }}
{{- else if eq .Kind "map" -}}
Kind: descriptor.TypeKindMap, Key: {{ template "typeDescriptor" .Key }}, Value: {{ template "typeDescriptor" .Value }}
{{- else if eq .Kind "enum" -}}
Kind: descriptor.TypeKindEnum, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}
{{- else if eq .Kind "data" -}}
Kind: descriptor.TypeKindData, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "config" -}}
Kind: descriptor.TypeKindConfig, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "event" -}}
Kind: descriptor.TypeKindEvent, Name: {{ quote .Name }}, SkelName: {{ quote .SkelName }}{{ template "typeArguments" .TypeArguments }}
{{- else if eq .Kind "typeParameter" -}}
Kind: descriptor.TypeKindTypeParameter, Name: {{ quote .Name }}
{{- end -}}
{{- if .Nullable }}, Nullable: true{{ end -}}
{{- end }}
