{{- define "domainSchema" -}}
var _DomainSchema = &skel.DomainSchema{
	Domain: {{ quote .Schema.Domain }},
	{{- if .Schema.Description }}
	Description: {{ quote .Schema.Description }},
	{{- end }}
	Hash: {{ quote .Schema.Hash }},
	Full: {{ .Schema.Full }},
	Generated: &skel.GeneratedInfo{CompilerVersion: {{ quote .Schema.Generated.CompilerVersion }}},
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
