package {{ .PackageName }}

import "go.yorun.ai/vine/core/skel"

func init() {
	skel.RegisterDomainSchema(_DomainSchema)
}

{{ template "domainSchema" . }}
