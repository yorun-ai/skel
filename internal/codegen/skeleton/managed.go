package skeleton

import (
	"go.yorun.ai/skel/internal/codegen/output"
	"go.yorun.ai/skel/internal/model"
)

// GenerateManaged commits generated files as one transaction.
func GenerateManaged(domain *model.Domain, option Option) error {
	return output.RunManagedOutputs([]string{option.Out}, func(staged []string) error {
		option.Out = staged[0]
		return Generate(domain, option)
	})
}
