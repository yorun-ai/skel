package {{ .PackageName }}

import (
	"go.yorun.ai/skel/descriptor"
	"go.yorun.ai/vine/core/skel"
)

func init() {
	skel.RegisterDomainDescriptor(_DomainDescriptor)
}

{{ template "domainDescriptor" . }}
