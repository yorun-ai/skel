package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/schema"
)

func TestGeneratorRendersPubGoView(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")
	goPubOutDir := filepath.Join(t.TempDir(), "skeledpub")

	userStatus := &schema.Enum{Pub: true, Name: "UserStatus", Items: []*schema.EnumItem{{Name: "ACTIVE"}}}
	unusedStatus := &schema.Enum{Pub: true, Name: "UnusedStatus", Items: []*schema.EnumItem{{Name: "IDLE"}}}
	publicStatus := &schema.Enum{Pub: true, Name: "PublicStatus", Items: []*schema.EnumItem{{Name: "READY"}}}
	address := &schema.Data{
		Pub:  true,
		Name: "Address",
		Members: []*schema.DataMember{
			{Name: "city", Type: codegentest.StringType()},
		},
	}
	user := &schema.Data{
		Pub:  true,
		Name: "User",
		Members: []*schema.DataMember{
			{Name: "status", Type: codegentest.EnumType(userStatus)},
			{Name: "address", Type: codegentest.DataType(address)},
		},
	}
	unusedData := &schema.Data{
		Pub:  true,
		Name: "UnusedData",
		Members: []*schema.DataMember{
			{Name: "idle", Type: codegentest.EnumType(unusedStatus)},
		},
	}
	partnerCredential := &schema.Data{
		Name: "PartnerActorCredential",
		Members: []*schema.DataMember{
			{Name: "subject", Type: codegentest.StringType()},
		},
	}
	partnerInfo := &schema.Data{
		Name: "PartnerActorInfo",
		Members: []*schema.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	publicCredential := &schema.Data{
		Name:      "PublicOnlyActorCredential",
		Sensitive: true,
		Members: []*schema.DataMember{
			{Name: "subject", Type: codegentest.StringType()},
		},
	}
	publicInfo := &schema.Data{
		Name:      "PublicOnlyActorInfo",
		Sensitive: true,
		Members: []*schema.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{
			{Pub: true, Name: "OpenAPIActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaAgent)}},
			{Name: "PartnerActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}, Auth: new(schema.ActorAuth{Credential: partnerCredential, Info: partnerInfo})},
			{Pub: true, Name: "PublicOnlyActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)}, Auth: new(schema.ActorAuth{Credential: publicCredential, Info: publicInfo})},
		},
		Enums: []*schema.Enum{userStatus, unusedStatus, publicStatus},
		Data:  []*schema.Data{address, user, unusedData},
		Configs: []*schema.Data{
			{
				Pub:       true,
				Name:      "DemoConfig",
				Lifecycle: schema.ConfigLifecycleEternal,
				Members: []*schema.DataMember{
					{Name: "status", Type: codegentest.EnumType(publicStatus)},
				},
			},
		},
		Services: []*schema.Service{
			{
				Pub:       true,
				Name:      "UserService",
				Audiences: []*schema.ActorAudience{{Actor: "OpenAPIActor"}},
				Methods: []*schema.Method{
					methodForTest("UserService", &schema.Method{Name: "getUser", ResultType: codegentest.DataType(user)}),
				},
			},
			{
				Name:      "PartnerService",
				Audiences: []*schema.ActorAudience{{Actor: "PartnerActor"}},
				Methods: []*schema.Method{
					methodForTest("PartnerService", &schema.Method{Name: "ping", ResultType: codegentest.StringType()}),
				},
			},
		},
		Events: []*schema.Data{
			{
				Pub:  true,
				Name: "UserCreatedEvent",
				Members: []*schema.DataMember{
					{Name: "user", Type: codegentest.DataType(user)},
				},
			},
			{
				Name: "PartnerEvent",
				Members: []*schema.DataMember{
					{Name: "message", Type: codegentest.StringType()},
				},
			},
		},
		Webs: []*schema.Web{
			{Name: "UserPortalWeb", Audiences: []*schema.ActorAudience{{Actor: "OpenAPIActor"}}},
		},
		Tasks: []*schema.Task{
			{
				Name: "RebuildUserIndexTask",
				Triggers: []*schema.TaskTrigger{
					triggerForTest("RebuildUserIndexTask", &schema.TaskTrigger{
						Name: "atTime",
						Arguments: []*schema.Argument{
							{Name: "startAt", Type: codegentest.LocalDateTimeType()},
						},
					}),
				},
			},
		},
	})

	if err := generateFixture(pkg, golang.Option{
		Out:          goOutDir,
		AsModule:     true,
		PubOut:       goPubOutDir,
		ModulePrefix: "github.com/acme/skel",
	}); err != nil {
		t.Fatal(err)
	}

	for _, outDir := range []string{goOutDir, goPubOutDir} {
		docContent := readFileForTest(t, filepath.Join(outDir, "doc.go"))
		if !strings.Contains(docContent, "// This package is fully managed by skelc.") ||
			!strings.Contains(docContent, "// unmanaged files.") {
			t.Fatalf("expected generated package ownership warning in %s/doc.go:\n%s", outDir, docContent)
		}
	}

	serviceContent, err := os.ReadFile(filepath.Join(goPubOutDir, "service.go"))
	if err != nil {
		t.Fatalf("read go service file: %v", err)
	}
	if !strings.Contains(string(serviceContent), "type UserServiceClient interface") {
		t.Fatalf("expected openapi service client, got:\n%s", string(serviceContent))
	}
	if strings.Contains(string(serviceContent), "type UserServiceServer interface") {
		t.Fatalf("did not expect pub-only codegen to render service server, got:\n%s", string(serviceContent))
	}
	if strings.Contains(string(serviceContent), "type PartnerServiceClient interface") {
		t.Fatalf("did not expect non-pub service in pub-only codegen, got:\n%s", string(serviceContent))
	}

	actorContent, err := os.ReadFile(filepath.Join(goPubOutDir, "actor.go"))
	if err != nil {
		t.Fatalf("read go actor file: %v", err)
	}
	if !strings.Contains(string(actorContent), "type OpenAPIActor struct {") {
		t.Fatalf("expected referenced pub actor, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "type PublicOnlyActor struct {") {
		t.Fatalf("expected explicitly pub actor, got:\n%s", string(actorContent))
	}
	if strings.Contains(string(actorContent), "type PartnerActor struct {") {
		t.Fatalf("did not expect non-pub actor in pub-only codegen, got:\n%s", string(actorContent))
	}
	if strings.Contains(string(actorContent), "PartnerActorAuthService") {
		t.Fatalf("did not expect non-pub actor auth service in pub-only codegen, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "rpc.Register(_PublicOnlyActorAuthServiceSpec)") {
		t.Fatalf("expected actor auth service to register, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "type PublicOnlyActorAuthServiceServer interface") {
		t.Fatalf("expected actor auth service server, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "Auth(credential PublicOnlyActorCredential) PublicOnlyActorInfo") {
		t.Fatalf("expected actor auth service auth method, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "type PublicOnlyActorCredential struct") || !strings.Contains(string(actorContent), "type PublicOnlyActorInfo struct") {
		t.Fatalf("expected actor credential/info data in actor.go, got:\n%s", string(actorContent))
	}
	if !strings.Contains(string(actorContent), "func (PublicOnlyActorCredential) SkelSensitive() {}") ||
		!strings.Contains(string(actorContent), "func (PublicOnlyActorInfo) SkelSensitive() {}") {
		t.Fatalf("expected sensitive actor credential/info marker methods, got:\n%s", string(actorContent))
	}
	if strings.Contains(string(actorContent), "type PublicOnlyActorAuthServiceClient interface") {
		t.Fatalf("did not expect actor auth service client, got:\n%s", string(actorContent))
	}
	if strings.Contains(string(actorContent), "type PublicOnlyActorAuthServiceClientER interface") {
		t.Fatalf("did not expect actor auth service er client, got:\n%s", string(actorContent))
	}

	dataContent, err := os.ReadFile(filepath.Join(goPubOutDir, "data.go"))
	if err != nil {
		t.Fatalf("read go data file: %v", err)
	}
	if strings.Contains(string(dataContent), "type PartnerActorCredential struct") || strings.Contains(string(dataContent), "type PartnerActorInfo struct") {
		t.Fatalf("did not expect non-pub actor credential/info data in pub-only go output, got:\n%s", string(dataContent))
	}
	if strings.Contains(string(dataContent), "type PublicOnlyActorCredential struct") || strings.Contains(string(dataContent), "type PublicOnlyActorInfo struct") {
		t.Fatalf("did not expect actor credential/info data in data.go, got:\n%s", string(dataContent))
	}

	eventContent, err := os.ReadFile(filepath.Join(goPubOutDir, "event.go"))
	if err != nil {
		t.Fatalf("read go event file: %v", err)
	}
	if !strings.Contains(string(eventContent), "type UserCreatedEventListener interface") {
		t.Fatalf("expected openapi event listener, got:\n%s", string(eventContent))
	}
	if strings.Contains(string(eventContent), "type UserCreatedEventEmitter interface") {
		t.Fatalf("did not expect pub-only codegen to render event emitter, got:\n%s", string(eventContent))
	}
	if strings.Contains(string(eventContent), "type PartnerEventListener interface") {
		t.Fatalf("did not expect non-pub event in pub-only codegen, got:\n%s", string(eventContent))
	}

	if !strings.Contains(string(dataContent), "type User struct") || !strings.Contains(string(dataContent), "type Address struct") {
		t.Fatalf("expected explicitly pub data, got:\n%s", string(dataContent))
	}
	if !strings.Contains(string(dataContent), "type UnusedData struct") {
		t.Fatalf("expected explicitly pub data in pub-only codegen, got:\n%s", string(dataContent))
	}

	enumContent, err := os.ReadFile(filepath.Join(goPubOutDir, "enum.go"))
	if err != nil {
		t.Fatalf("read go enum file: %v", err)
	}
	if !strings.Contains(string(enumContent), "type UserStatus string") {
		t.Fatalf("expected explicitly pub enum, got:\n%s", string(enumContent))
	}
	if !strings.Contains(string(enumContent), "type UnusedStatus string") {
		t.Fatalf("expected explicitly pub enum, got:\n%s", string(enumContent))
	}
	if !strings.Contains(string(enumContent), "type PublicStatus string") {
		t.Fatalf("expected explicitly pub enum, got:\n%s", string(enumContent))
	}

	configContent, err := os.ReadFile(filepath.Join(goPubOutDir, "config.go"))
	if err != nil {
		t.Fatalf("read go config file: %v", err)
	}
	if !strings.Contains(string(configContent), "type DemoConfig struct") {
		t.Fatalf("expected explicitly pub config, got:\n%s", string(configContent))
	}
	schemaContent, err := os.ReadFile(filepath.Join(goPubOutDir, "descriptor.go"))
	if err != nil {
		t.Fatalf("read go schema file: %v", err)
	}
	codegentest.AssertGoSourceContains(t, string(schemaContent), `Name: "demo.user"`)
	assertFileMissing(t, filepath.Join(goPubOutDir, "task.go"))
}

func TestGeneratorIncludesImplicitPubDependencies(t *testing.T) {
	goOutDir := filepath.Join(t.TempDir(), "skeled")
	goPubOutDir := filepath.Join(t.TempDir(), "skeledpub")

	user := &schema.Data{
		Name: "User",
		Members: []*schema.DataMember{
			{Name: "name", Type: codegentest.StringType()},
		},
	}
	pkg := newSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.user",
		Actors: []*schema.Actor{
			{Pub: true, Name: "OpenAPIActor", Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaAgent)}},
		},
		Data: []*schema.Data{user},
		Services: []*schema.Service{
			{
				Pub:       true,
				Name:      "UserService",
				Audiences: []*schema.ActorAudience{{Actor: "OpenAPIActor"}},
				Methods: []*schema.Method{
					methodForTest("UserService", &schema.Method{Name: "getUser", ResultType: codegentest.DataType(user)}),
				},
			},
		},
	})

	err := generateFixture(pkg, golang.Option{
		Out:          goOutDir,
		AsModule:     true,
		PubOut:       goPubOutDir,
		ModulePrefix: "github.com/acme/skel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content := readFileForTest(t, filepath.Join(goPubOutDir, "data.go")); !strings.Contains(content, "type User struct") {
		t.Fatalf("missing implicit dependency: %s", content)
	}
}
