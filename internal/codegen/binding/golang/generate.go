package golang

import (
	"context"
	"fmt"

	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/schema"
)

// NewGenerator constructs a generator without rendering or writing files.
// Out supplies the intended location for target-specific package naming only.
// Returned files use the primary target ""; split Go public files use "pub".
// The generator selects its configured surface from Input's complete graph.
func NewGenerator(option Option) (codegen.Generator, error) {
	resolved, err := NormalizeOption(option)
	if err != nil {
		return nil, err
	}
	return generator(resolved), nil
}
func generator(option ResolvedOption) codegen.Generator {
	return codegen.GeneratorFunc(func(ctx context.Context, input codegen.Input) ([]codegen.File, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !input.Valid() {
			return nil, fmt.Errorf("codegen input is uninitialized")
		}
		settings := option.Options()
		if err := validateGolangImports(input.Schema(), settings); err != nil {
			return nil, err
		}
		files := new(binding.FileCollector{Context: ctx})
		if err := render(input, option, files); err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return files.Files, nil
	})
}

// Generate uses the shared SDK preparation and output transaction.
func Generate(domain *schema.Domain, option ResolvedOption) error {
	input, err := codegen.Prepare(domain, codegen.Selection{})
	if err != nil {
		return err
	}
	settings := option.Options()
	targets := map[string]string{"": settings.Out}
	if settings.PubOut != "" {
		targets["pub"] = settings.PubOut
	}
	return codegen.Run(context.Background(), input, generator(option), targets)
}
