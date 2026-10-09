package source

import (
	"go.yorun.ai/skel/schema"
	"go/token"
	"testing"
)

func TestBuildMethodNames(t *testing.T) {
	names := buildMethodNames([]*MethodArgument{
		{Name: "client"},
		{Name: "ctx"},
		{Name: "ret"},
		{Name: "err"},
	}, nil)

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

func TestBuildMethodNamesWithoutCollisions(t *testing.T) {
	names := buildMethodNames([]*MethodArgument{{Name: "value"}}, nil)

	if names.ReceiverName != "client" || names.ContextName != "ctx" || names.ResultName != "ret" || names.ErrorName != "err" || names.OptionsName != "_ivOpts" {
		t.Fatalf("unexpected names: %+v", names)
	}
}

func TestEscapeArgumentNames(t *testing.T) {
	for keyword := token.BREAK; keyword <= token.VAR; keyword++ {
		if !keyword.IsKeyword() {
			continue
		}
		name := keyword.String()
		args := []*MethodArgument{{Name: name, SkelName: name}, {Name: name + "_"}}
		escapeArgumentNames(args, nil)
		if args[0].Name != name+"__" || args[0].SkelName != name {
			t.Fatalf("keyword %q: %+v", name, args[0])
		}
	}
	args := []*MethodArgument{
		{Name: "Type"}, {Name: "typeName"}, {Name: "recover"}, {Name: "ex"},
		{Name: "skeltype"}, {Name: "string"}, {Name: "other"},
	}
	escapeArgumentNames(args, &Type{Plain: "[]string"})
	want := []string{"Type", "typeName", "recover_", "ex_", "skeltype_", "string_", "other"}
	for i, argument := range args {
		if argument.Name != want[i] {
			t.Errorf("argument %d = %q, want %q", i, argument.Name, want[i])
		}
	}
	member := (_Types{}).castDataMember(&schema.DataMember{Name: "type", Type: &schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}})
	if member.Name != "Type" || member.SkelName != "type" {
		t.Fatalf("field = %+v", member)
	}
}

func TestBuildMethodNamesServerAndTaskCollisions(t *testing.T) {
	names := buildMethodNames([]*MethodArgument{{Name: "service"}, {Name: "launcher"}, {Name: "runner"}, {Name: "ret"}, {Name: "err"}}, nil)
	if names.ServerReceiverName != "service_" || names.LauncherReceiverName != "launcher_" || names.RunnerReceiverName != "runner_" || names.ResultName != "ret_" || names.ErrorName != "err_" {
		t.Fatalf("unexpected names: %+v", names)
	}
}

func TestBuildMethodNamesAvoidsTypeImports(t *testing.T) {
	names := buildMethodNames(nil, &Type{Plain: "client.Result[ret.Value]"})
	if names.ReceiverName != "client_" || names.ResultName != "ret_" {
		t.Fatalf("names shadow type imports: %+v", names)
	}
}
