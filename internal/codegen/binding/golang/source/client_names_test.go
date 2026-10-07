package source

import "testing"

func TestBuildClientMethodNames(t *testing.T) {
	names := buildClientMethodNames([]*MethodArgument{
		{Name: "client"},
		{Name: "ctx"},
		{Name: "ret"},
		{Name: "err"},
	})

	if names.ReceiverName != "client_" {
		t.Fatalf("receiver = %q, want client_", names.ReceiverName)
	}
	if names.ContextName != "ctx_" {
		t.Fatalf("context = %q, want ctx_", names.ContextName)
	}
	if names.ResultName != "ret_" {
		t.Fatalf("result = %q, want ret_", names.ResultName)
	}
	if names.ErrorName != "err_" {
		t.Fatalf("error = %q, want err_", names.ErrorName)
	}
	if names.OptionsName != "_ivOpts" {
		t.Fatalf("options = %q, want _ivOpts", names.OptionsName)
	}
}

func TestBuildClientMethodNamesWithoutCollisions(t *testing.T) {
	names := buildClientMethodNames([]*MethodArgument{{Name: "value"}})

	if names.ReceiverName != "client" || names.ContextName != "ctx" || names.ResultName != "ret" || names.ErrorName != "err" || names.OptionsName != "_ivOpts" {
		t.Fatalf("unexpected names: %+v", names)
	}
}
