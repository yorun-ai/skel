// Package typescript generates TypeScript source and module metadata from a
// validated semantic model. Rendering payloads and import state stay local.
package typescript

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/internal/codegen/binding"
)

func render(validated codegen.Input, option Option, sink binding.FileSink) error {
	result, err := generateSource(validated, option.Out, option, sink)
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
			Sink:            sink,
			PackageName:     result.PackageName,
			Imports:         option.Imports,
			ResolvedImports: result.ResolvedImports,
		})
	}
	return nil
}
