package typescript

import (
	"fmt"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/codegen/typescript/module"
	"go.yorun.ai/skelc/internal/codegen/typescript/source"
	"go.yorun.ai/skelc/internal/model"
)

func Generate(domain *model.Domain, option Option) error {
	validated, err := common.PrepareDomain(domain)
	if err != nil {
		return fmt.Errorf("validate TypeScript generation model: %w", err)
	}
	result, err := source.GenerateValidated(validated, option.Out, source.Option{
		ModuleScope: option.ModuleScope,
		Module:      option.Module,
		Imports:     option.Imports,
	})
	if err != nil {
		return err
	}
	if option.AsModule {
		imports := map[string]string{}
		for domain := range result.ResolvedImports {
			if path := option.Imports[domain]; path != "" {
				imports[domain] = path
			}
		}
		option.Imports = imports
		return module.Generate(module.Option{
			Out:             option.Out,
			PackageName:     result.PackageName,
			Imports:         option.Imports,
			ResolvedImports: result.ResolvedImports,
		})
	}
	return nil
}
