package golang_test

import (
	"go.yorun.ai/skel/internal/codegen/golang"
	"go.yorun.ai/skel/internal/model"
)

func generateFixture(domain *model.Domain, option golang.Option) error {
	if option.CompilerVersion == "" {
		option.CompilerVersion = "v0.0.0-dev"
	}
	resolved, err := golang.ResolveOption(option)
	if err != nil {
		return err
	}
	return golang.Generate(domain, resolved)
}
