package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
)

func TestGeneratorAlwaysRendersGoDocFile(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	pkg := newModelDomainForTest(t, codegentest.DomainModel("demo.user"))

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goDocContent, err := os.ReadFile(filepath.Join(goOutDir, "doc.go"))
	if err != nil {
		t.Fatalf("read go doc file: %v", err)
	}
	if !strings.Contains(string(goDocContent), "// Package skeled") {
		t.Fatalf("expected fallback go doc comment, got:\n%s", string(goDocContent))
	}
	if !strings.Contains(string(goDocContent), "// This package is fully managed by skelc.") ||
		!strings.Contains(string(goDocContent), "// unmanaged files.") {
		t.Fatalf("expected generated package ownership warning, got:\n%s", string(goDocContent))
	}
	if !strings.Contains(string(goDocContent), "package skeled") {
		t.Fatalf("expected go doc package declaration, got:\n%s", string(goDocContent))
	}
	if strings.Contains(string(goDocContent), "\n\npackage skeled") {
		t.Fatalf("did not expect blank line between fallback doc comment and package declaration, got:\n%s", string(goDocContent))
	}

	assertFileMissing(t, filepath.Join(goOutDir, "go.mod"))
	assertFileMissing(t, filepath.Join(goOutDir, "go.sum"))
}

func TestGeneratorRendersDescriptionComments(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")

	userStatus := &model.Enum{
		Name:        "UserStatus",
		Description: "User status",
		Items: []*model.EnumItem{
			{Name: "ACTIVE", Description: "Active"},
		},
	}
	userProfile := &model.Data{
		Name:        "UserProfile",
		Description: "User profile",
		Sensitive:   true,
		Members: []*model.DataMember{
			{
				Name:        "avatarUrl",
				Description: "Avatar URL",
				Example:     `"https://xxx.com/a.png"`,
				Sensitive:   true,
				Type:        codegentest.NullableType(codegentest.StringType()),
			},
			{
				Name:        "status",
				Description: "User status",
				Type:        codegentest.EnumType(userStatus),
			},
		},
	}
	pkg := newModelDomainForTest(t, model.DomainSpec{
		Name:        "demo.user",
		Description: "User domain",
		Enums:       []*model.Enum{userStatus},
		Data:        []*model.Data{userProfile},
		Actors: []*model.Actor{
			{Name: "ClientActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaClient)}},
			{Name: "PartnerActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaAgent)}},
			{Name: "OpenAPIActor", Vias: []*model.ActorVia{codegentest.ActorVia(model.ActorViaOpenAPI)}},
		},
		Services: []*model.Service{
			{
				Name:        "UserService",
				Description: "User service",
				Audiences:   []*model.ActorAudience{{Actor: "ClientActor"}},
				Methods: []*model.Method{
					methodForTest("UserService", &model.Method{
						Name:               "getUser",
						Description:        "Get a user by ID",
						InputDescription:   "Input parameters",
						ArgumentsSensitive: true,
						OutputDescription:  "User information",
						OutputExample:      `{ id:10001, avatarUrl:"https://xxx.com/a.png" }`,
						ResultSensitive:    true,
						ResultType:         codegentest.DataType(userProfile),
						Arguments: []*model.Argument{
							{
								Name:        "userId",
								Description: "User ID",
								Example:     `"10001"`,
								Sensitive:   true,
								Type:        codegentest.StringType(),
							},
						},
					}),
				},
			},
		},
		Tasks: []*model.Task{
			{
				Name:        "RebuildUserIndexTask",
				Description: "Rebuild the user index",
				Triggers: []*model.TaskTrigger{
					triggerForTest("RebuildUserIndexTask", &model.TaskTrigger{
						Name:               "atTime",
						Description:        "Scheduled trigger",
						InputDescription:   "Trigger parameters",
						ArgumentsSensitive: true,
						Arguments: []*model.Argument{
							{
								Name:        "startAt",
								Description: "Start time",
								Example:     `"2026-05-04T12:00:00"`,
								Sensitive:   true,
								Type:        codegentest.LocalDateTimeType(),
							},
						},
					}),
				},
			},
		},
		Configs: []*model.Data{
			{
				Name:      "SecretConfig",
				Kind:      model.DataKindConfig,
				Lifecycle: model.ConfigLifecycleEternal,
				Sensitive: true,
				Members: []*model.DataMember{
					{Name: "token", Type: codegentest.StringType()},
				},
			},
		},
		Events: []*model.Data{
			{
				Name:        "UserCreatedEvent",
				Description: "User created event",
				Sensitive:   true,
				Members: []*model.DataMember{
					{Name: "userId", Description: "User ID", Sensitive: true, Type: codegentest.StringType()},
				},
			},
		},
	})

	if err := generateFixture(pkg, golang.Option{Out: goOutDir}); err != nil {
		t.Fatal(err)
	}

	goDocContent, err := os.ReadFile(filepath.Join(goOutDir, "doc.go"))
	if err != nil {
		t.Fatalf("read go doc file: %v", err)
	}
	if !strings.Contains(string(goDocContent), "package skeled") {
		t.Fatalf("expected go doc package declaration, got:\n%s", string(goDocContent))
	}
	if !strings.Contains(string(goDocContent), "// Package skeled User domain") {
		t.Fatalf("expected go doc comment from domain description, got:\n%s", string(goDocContent))
	}
	if strings.HasPrefix(string(goDocContent), "\n") {
		t.Fatalf("did not expect leading blank line in doc.go, got:\n%s", string(goDocContent))
	}
	if strings.Contains(string(goDocContent), "\n\npackage skeled") {
		t.Fatalf("did not expect blank line between doc comment and package declaration, got:\n%s", string(goDocContent))
	}

	assertFileMissing(t, filepath.Join(goOutDir, "go.mod"))
	assertFileMissing(t, filepath.Join(goOutDir, "go.sum"))

	goServiceContent, err := os.ReadFile(filepath.Join(goOutDir, "service.go"))
	if err != nil {
		t.Fatalf("read go service file: %v", err)
	}
	if !strings.Contains(string(goServiceContent), "// UserServiceServer User service") {
		t.Fatalf("expected go service comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "// GetUser Get a user by ID.") {
		t.Fatalf("expected go method comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "//   @param userId - User ID (e.g. \"10001\")") {
		t.Fatalf("expected go method argument comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), `json:"userId" skel:"index(0),sensitive"`) {
		t.Fatalf("expected sensitive service argument tag, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "//   @returns UserProfile - User information (e.g. { id:10001, avatarUrl:\"https://xxx.com/a.png\" })") {
		t.Fatalf("expected go method return comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "type UserServiceClient interface {\n\t// GetUser Get a user by ID.") {
		t.Fatalf("expected go client interface method comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "type UserServiceClientER interface {\n\t// GetUser Get a user by ID.") {
		t.Fatalf("expected go er client interface method comment, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "MethodFuncs: []any{") {
		t.Fatalf("expected go service method function metadata, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "ArgumentsSensitive:          true,") ||
		!strings.Contains(string(goServiceContent), "ResultSensitive:             true,") {
		t.Fatalf("expected whole input and output sensitive specs, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "UserServiceClient.GetUser") {
		t.Fatalf("expected go service method function metadata to include client method, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "UserServiceClientER.GetUser") {
		t.Fatalf("expected go service method function metadata to include er client method, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "UserServiceServer.GetUser") {
		t.Fatalf("expected go service method function metadata to include server method, got:\n%s", string(goServiceContent))
	}
	if !strings.Contains(string(goServiceContent), "UserServiceServerER.GetUser") {
		t.Fatalf("expected go service method function metadata to include er server method, got:\n%s", string(goServiceContent))
	}
	if strings.Contains(string(goServiceContent), "_UserServiceGetUserArguments Input parameters") {
		t.Fatalf("did not expect go arguments data comment, got:\n%s", string(goServiceContent))
	}
	if strings.Contains(string(goServiceContent), "// UserId User ID") {
		t.Fatalf("did not expect go argument field comment, got:\n%s", string(goServiceContent))
	}

	goEnumContent, err := os.ReadFile(filepath.Join(goOutDir, "enum.go"))
	if err != nil {
		t.Fatalf("read go enum file: %v", err)
	}
	if !strings.Contains(string(goEnumContent), "// UserStatus User status") {
		t.Fatalf("expected go enum comment, got:\n%s", string(goEnumContent))
	}
	if !strings.Contains(string(goEnumContent), "// UserStatusActive Active") {
		t.Fatalf("expected go enum item comment, got:\n%s", string(goEnumContent))
	}

	goDataContent, err := os.ReadFile(filepath.Join(goOutDir, "data.go"))
	if err != nil {
		t.Fatalf("read go data file: %v", err)
	}
	if !strings.Contains(string(goDataContent), "// UserProfile User profile") {
		t.Fatalf("expected go data comment, got:\n%s", string(goDataContent))
	}
	if !strings.Contains(string(goDataContent), `// AvatarUrl Avatar URL (e.g. "https://xxx.com/a.png")`) {
		t.Fatalf("expected go data member merged comment, got:\n%s", string(goDataContent))
	}
	if !strings.Contains(string(goDataContent), `json:"avatarUrl" skel:"sensitive"`) {
		t.Fatalf("expected sensitive data member tag, got:\n%s", string(goDataContent))
	}
	if !strings.Contains(string(goDataContent), "func (UserProfile) SkelSensitive() {}") {
		t.Fatalf("expected sensitive data marker method, got:\n%s", string(goDataContent))
	}

	goActorContent, err := os.ReadFile(filepath.Join(goOutDir, "actor.go"))
	if err != nil {
		t.Fatalf("read go actor file: %v", err)
	}
	if !strings.Contains(string(goActorContent), `type ClientActor struct {`) {
		t.Fatalf("expected go actor type, got:\n%s", string(goActorContent))
	}
	if !strings.Contains(string(goActorContent), `func (ClientActor) SkelName() string {`) {
		t.Fatalf("expected go actor skel name method, got:\n%s", string(goActorContent))
	}
	if !strings.Contains(string(goActorContent), `return "demo.user.ClientActor"`) {
		t.Fatalf("expected go actor skel name, got:\n%s", string(goActorContent))
	}
	if !strings.Contains(string(goActorContent), "return []skel.ActorVia{\n\t\tskel.ActorViaClient,") {
		t.Fatalf("expected go actor vias, got:\n%s", string(goActorContent))
	}

	goSchemaContent, err := os.ReadFile(filepath.Join(goOutDir, "schema.go"))
	if err != nil {
		t.Fatalf("read go schema file: %v", err)
	}
	if !strings.Contains(string(goSchemaContent), `Hash:`) {
		t.Fatalf("expected go schema hash fields, got:\n%s", string(goSchemaContent))
	}
	if !strings.Contains(string(goSchemaContent), "ArgumentsSensitive:") ||
		!strings.Contains(string(goSchemaContent), "ResultSensitive:") ||
		!strings.Contains(string(goSchemaContent), "Sensitive:") {
		t.Fatalf("expected sensitive metadata in go schema, got:\n%s", string(goSchemaContent))
	}
	initIndex := strings.Index(string(goSchemaContent), "func init()")
	schemaIndex := strings.Index(string(goSchemaContent), "var _DomainSchema")
	if initIndex < 0 || schemaIndex < 0 || initIndex > schemaIndex {
		t.Fatalf("expected schema init before schema var, got:\n%s", string(goSchemaContent))
	}

	goTaskContent, err := os.ReadFile(filepath.Join(goOutDir, "task.go"))
	if err != nil {
		t.Fatalf("read go task file: %v", err)
	}
	if !strings.Contains(string(goTaskContent), "// RebuildUserIndexTaskRunner Rebuild the user index") {
		t.Fatalf("expected go task comment, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "// AtTime Scheduled trigger.") {
		t.Fatalf("expected go task trigger comment, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "//   @param startAt - Start time (e.g. \"2026-05-04T12:00:00\")") {
		t.Fatalf("expected go task param comment, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), `json:"startAt" skel:"sensitive"`) {
		t.Fatalf("expected sensitive task argument tag, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "type RebuildUserIndexTaskLauncher interface {\n\t// LaunchAtTime Scheduled trigger.") {
		t.Fatalf("expected go task launcher method comment, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "type RebuildUserIndexTaskRunner interface {\n\t// RunAtTime Scheduled trigger.") {
		t.Fatalf("expected go task runner method comment, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "Name:               \"AtTime\",") {
		t.Fatalf("expected go task trigger name to keep original value, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "LauncherMethodName: \"LaunchAtTime\",") {
		t.Fatalf("expected go task trigger launcher method name, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "RunnerMethodName:   \"RunAtTime\",") {
		t.Fatalf("expected go task trigger runner method name, got:\n%s", string(goTaskContent))
	}
	if !strings.Contains(string(goTaskContent), "ArgumentsSensitive: true,") {
		t.Fatalf("expected sensitive task trigger input metadata, got:\n%s", string(goTaskContent))
	}

	goEventContent, err := os.ReadFile(filepath.Join(goOutDir, "event.go"))
	if err != nil {
		t.Fatalf("read go event file: %v", err)
	}
	if !strings.Contains(string(goEventContent), `json:"userId" skel:"sensitive"`) {
		t.Fatalf("expected sensitive event member tag, got:\n%s", string(goEventContent))
	}
	if !strings.Contains(string(goEventContent), "func (UserCreatedEvent) SkelSensitive() {}") {
		t.Fatalf("expected sensitive event payload marker method, got:\n%s", string(goEventContent))
	}
	if !strings.Contains(string(goEventContent), `EmitterMethodName:`) ||
		!strings.Contains(string(goEventContent), `"EmitUserCreated",`) {
		t.Fatalf("expected go event emitter method name, got:\n%s", string(goEventContent))
	}

	goConfigContent, err := os.ReadFile(filepath.Join(goOutDir, "config.go"))
	if err != nil {
		t.Fatalf("read go config file: %v", err)
	}
	if !strings.Contains(string(goConfigContent), "func (SecretConfig) SkelSensitive() {}") {
		t.Fatalf("expected sensitive config marker method, got:\n%s", string(goConfigContent))
	}
}
