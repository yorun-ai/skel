package typescript

import (
	"go.yorun.ai/skel/internal/codegen/output"
	"go.yorun.ai/skel/internal/model"
)

// GenerateManaged validates model imports and commits generated files as one transaction.
func GenerateManaged(domain *model.Domain, option Option) error {
	if err := validateTypeScriptImports(domain, option); err != nil {
		return err
	}
	return output.RunManagedOutputs([]string{option.Out}, func(staged []string) error {
		option.Out = staged[0]
		return Generate(domain, option)
	})
}
