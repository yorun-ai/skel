{{- define "domainDescriptor" -}}
var _DomainDescriptor = &descriptor.Domain{
	Name: {{ quote .Descriptor.Name }},
	{{- if .Descriptor.Description }}
	Description: {{ quote .Descriptor.Description }},
	{{- end }}
	Hash: {{ quote .Descriptor.Hash }},
	Full: {{ .Descriptor.Full }},
	Generated: &descriptor.GeneratedInfo{CompilerVersion: {{ quote .Descriptor.Generated.CompilerVersion }}},
	{{ template "domainEnums" . }}
	{{ template "domainData" . }}
	{{ template "domainConfigs" . }}
	{{ template "domainWebs" . }}
	{{ template "domainEvents" . }}
	{{ template "domainActors" . }}
	{{ template "domainResources" . }}
	{{ template "domainServices" . }}
	{{ template "domainTasks" . }}
}
{{- end }}
