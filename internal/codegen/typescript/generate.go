// Package typescript generates TypeScript source and module metadata from a
// validated semantic model. Rendering payloads and import state stay local.
package typescript

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
)

func Generate(domain *model.Domain, option Option) error {
	validated, err := common.PrepareDomain(domain)
	if err != nil {
		return fmt.Errorf("validate TypeScript generation model: %w", err)
	}
	result, err := generateSource(validated, option.Out, option)
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
		return generateModule(_ModuleOption{
			Out:             option.Out,
			PackageName:     result.PackageName,
			Imports:         option.Imports,
			ResolvedImports: result.ResolvedImports,
		})
	}
	return nil
}
