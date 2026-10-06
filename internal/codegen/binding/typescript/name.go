package typescript

import (
	"fmt"
	"strings"
)

func buildPackageName(moduleScope string, domainName string) string {
	parts := strings.Split(domainName, ".")
	parts[len(parts)-1] += "api"
	name := strings.Join(parts, "-")
	if strings.Contains(strings.TrimPrefix(moduleScope, "@"), "/") {
		return fmt.Sprintf("%s-%s", moduleScope, name)
	}
	return fmt.Sprintf("%s/%s", moduleScope, name)
}

func importPackageAlias(domainName string) string {
	parts := strings.Split(domainName, ".")
	return parts[len(parts)-1] + "api"
}
