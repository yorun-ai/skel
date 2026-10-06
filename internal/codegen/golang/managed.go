package golang

import (
	"go.yorun.ai/skel/internal/codegen/output"
	"go.yorun.ai/skel/internal/model"
)

// GenerateManaged validates model imports and commits generated files as one transaction.
func GenerateManaged(domain *model.Domain, resolved ResolvedOption) error {
	option := resolved.Options()
	if err := validateGolangImports(domain, option); err != nil {
		return err
	}
	return output.RunManagedOutputs([]string{option.Out, option.PubOut}, func(staged []string) error {
		return Generate(domain, resolved.WithOutputs(staged[0], staged[1]))
	})
}
