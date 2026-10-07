package golang

import (
	"errors"
	"testing"

	"go.yorun.ai/skel/internal/optionvalidation"
)

func TestResolveOptionSelectsRuntime(t *testing.T) {
	backend, err := ResolveOption(Option{CompilerVersion: developmentCompilerVersion})
	if err != nil || backend.Options().VineVersion != DefaultVineVersion || backend.Options().VrpcVersion != "" {
		t.Fatalf("backend: %+v, %v", backend, err)
	}
	client, err := ResolveOption(Option{ApiOnly: true})
	if err != nil || client.Options().VrpcVersion != "v0.13.0" || client.Options().VineVersion != "" {
		t.Fatalf("client: %+v, %v", client, err)
	}
	for _, option := range []Option{{VrpcVersion: "v0.13.0"}, {ApiOnly: true, VineVersion: "v0.28.0"}} {
		if _, err := ResolveOption(option); err == nil {
			t.Fatalf("accepted unused version: %+v", option)
		}
	}
	_, err = ResolveOption(Option{VineVersion: "v0.18.9"})
	var validation *optionvalidation.ValidationError
	if !errors.As(err, &validation) || validation.Field != optionvalidation.FieldGoVineVersion {
		t.Fatalf("lost structured error: %v", err)
	}
}

func TestCompilerVersionBaseline(t *testing.T) {
	for _, version := range []string{"", "v0.14.0", "v0.17.0", "v0.17.1-rc.1", "v0.0.0-dev.1", "invalid"} {
		_, err := ResolveOption(Option{CompilerVersion: version})
		var validation *optionvalidation.ValidationError
		if !errors.As(err, &validation) || validation.Field != optionvalidation.FieldGoCompilerVersion {
			t.Fatalf("accepted unsupported compiler %q: %v", version, err)
		}
	}
	for _, version := range []string{"v0.17.1", "v0.19.2", "v0.0.0-dev"} {
		if _, err := ResolveOption(Option{CompilerVersion: version}); err != nil {
			t.Fatalf("rejected compiler %s: %v", version, err)
		}
	}
	if _, err := ResolveOption(Option{ApiOnly: true}); err != nil {
		t.Fatalf("API client required unused compiler metadata: %v", err)
	}
}
