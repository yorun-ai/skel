package source

import (
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/schema"
)

func TestBuildServiceNames(t *testing.T) {
	names := buildServiceNames("userService")

	if names.Name != "UserService" {
		t.Fatalf("unexpected service name: %s", names.Name)
	}
	if names.ServerName != "UserServiceServer" {
		t.Fatalf("unexpected server name: %s", names.ServerName)
	}
	if names.ClientCtorName != "NewUserServiceClient" {
		t.Fatalf("unexpected client ctor name: %s", names.ClientCtorName)
	}
	if names.ERClientName != "UserServiceClientER" {
		t.Fatalf("unexpected er client name: %s", names.ERClientName)
	}
}

func TestCastService(t *testing.T) {
	service := new(_Gen).castService(&schema.Service{
		Name:        "UserService",
		SkelName:    "demo.user.UserService",
		Description: "User service",
		Methods: []*schema.Method{
			{
				Name:               "getUser",
				Description:        "Get a user by ID",
				ArgumentsSensitive: true,
				ResultSensitive:    true,
				Arguments: []*schema.Argument{
					{
						Name:        "userId",
						Description: "User ID",
						Example:     `"10001"`,
						Type: &schema.Type{
							Kind:   schema.TypeKindScalar,
							Scalar: schema.ScalarInt,
						},
					},
				},
				ResultType: &schema.Type{
					Kind: schema.TypeKindData,
					Data: &schema.Data{
						Name: "User",
					},
					Nullable: true,
				},
				OutputDescription: "User information",
				OutputExample:     `{ id:10001, name:"zhangsan" }`,
				ArgumentsData: &schema.Data{
					Name: "UserServiceGetUserArguments",
					Members: []*schema.DataMember{
						{
							Name: "userId",
							Type: &schema.Type{
								Kind:   schema.TypeKindScalar,
								Scalar: schema.ScalarInt,
							},
						},
					},
				},
			},
		},
	}, false, false)

	if len(service.CommentLines) == 0 || service.CommentLines[0] != "UserServiceServer User service" {
		t.Fatalf("unexpected service comment lines: %+v", service.CommentLines)
	}
	if len(service.Methods) != 1 {
		t.Fatalf("unexpected method count: %d", len(service.Methods))
	}
	if got := service.Methods[0].CommentLines; !reflect.DeepEqual(got, []string{
		"GetUser Get a user by ID.",
		`@param userId - User ID (e.g. "10001")`,
		`@returns *User - User information (e.g. { id:10001, name:"zhangsan" })`,
	}) {
		t.Fatalf("unexpected method comment lines: %+v", got)
	}
	if service.Methods[0].ArgumentsData == nil || service.Methods[0].ArgumentsData.Name != "_UserServiceGetUserArguments" {
		t.Fatalf("unexpected arguments data: %+v", service.Methods[0].ArgumentsData)
	}
	if !service.Methods[0].ArgumentsSensitive || !service.Methods[0].ResultSensitive {
		t.Fatal("expected whole input and output sensitive flags")
	}
	if service.Methods[0].Arguments[0].MemberName != "UserId" {
		t.Fatalf("unexpected argument member name: %s", service.Methods[0].Arguments[0].MemberName)
	}
	if service.Methods[0].ArgumentsContainsBinaryType {
		t.Fatalf("expected method arguments to not contain binary type")
	}
	if service.Methods[0].ResultContainsBinaryType {
		t.Fatalf("expected method result to not contain binary type")
	}
}

func TestGeneratedGoParsesWithMultilineDeprecatedReasons(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Services: []*schema.Service{{
			Name:             "UserService",
			Deprecated:       true,
			DeprecatedReason: "Use ProfileService instead\nComplete migration first",
			Methods: []*schema.Method{{
				Name:             "getUser",
				Deprecated:       true,
				DeprecatedReason: "Use getProfile instead\nThe old response will be removed",
			}},
		}},
		Tasks: []*schema.Task{{
			Name:             "RefreshTask",
			Deprecated:       true,
			DeprecatedReason: "Use RebuildTask instead\nThe old schedule will be removed",
			Triggers: []*schema.TaskTrigger{{
				Name:             "legacy",
				Deprecated:       true,
				DeprecatedReason: "Use manually instead\nLegacy scheduling is disabled",
			}},
		}},
	})
	outputDir := t.TempDir()
	gen := newGen(Option{
		Domain: pkg, View: mustView(t, view.ModeFull, pkg), Mode: view.ModeFull,
		PackageName: "skeled", Out: outputDir,
	})
	gen.genServiceGo()
	gen.genTaskGo()

	for _, filename := range []string{serviceGoFilename, taskGoFilename} {
		path := filepath.Join(outputDir, filename)
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, goparser.ParseComments)
		if err != nil {
			t.Fatalf("generated Go should parse: %v", err)
		}
		if file.Name.Name != "skeled" {
			t.Fatalf("expected package skeled in %s, got %s", path, file.Name.Name)
		}
	}
}

func TestCastServiceMarksMethodBinaryFlags(t *testing.T) {
	service := new(_Gen).castService(&schema.Service{
		Name:     "AssetService",
		SkelName: "demo.asset.AssetService",
		Methods: []*schema.Method{
			{
				Name: "upload",
				Arguments: []*schema.Argument{
					{
						Name: "payload",
						Type: &schema.Type{
							Kind:   schema.TypeKindScalar,
							Scalar: schema.ScalarBinary,
						},
					},
				},
				ArgumentsData: &schema.Data{
					Name: "AssetServiceUploadArguments",
					Members: []*schema.DataMember{
						{
							Name: "payload",
							Type: &schema.Type{
								Kind:   schema.TypeKindScalar,
								Scalar: schema.ScalarBinary,
							},
						},
					},
				},
			},
			{
				Name: "download",
				ResultType: &schema.Type{
					Kind:   schema.TypeKindScalar,
					Scalar: schema.ScalarBinary,
				},
			},
		},
	}, false, false)

	if !service.Methods[0].ArgumentsContainsBinaryType {
		t.Fatalf("expected upload method arguments to contain binary type")
	}
	if service.Methods[0].ResultContainsBinaryType {
		t.Fatalf("expected upload method result to not contain binary type")
	}
	if service.Methods[1].ArgumentsContainsBinaryType {
		t.Fatalf("expected download method arguments to not contain binary type")
	}
	if !service.Methods[1].ResultContainsBinaryType {
		t.Fatalf("expected download method result to contain binary type")
	}
}

func TestBuildServiceImports(t *testing.T) {
	imports := buildServiceImports([]*Service{
		{
			Methods: []*ServiceMethod{
				{
					ResultType: &Type{
						Imports: []*Import{{Path: skelImport}},
					},
					Arguments: []*MethodArgument{
						{
							Type: &Type{
								Imports: []*Import{{Path: skelImport}},
							},
						},
					},
				},
			},
		},
	})

	if got, want := importPaths(imports), []string{"go.yorun.ai/vine/core/ex", "go.yorun.ai/vine/core/rpc", skelImport, "reflect"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected import paths: got=%v want=%v", got, want)
	}
}
