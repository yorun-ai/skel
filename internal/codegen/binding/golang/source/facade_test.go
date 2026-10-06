package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/schema"
)

func TestFacadeGoUsesSelectedPublicContract(t *testing.T) {
	const privateSource = `domain demo
enum State { READY }
data Payload { state: State }
data Local { value: string }
config LocalConfig eternal { value: string }
actor LocalActor { via client {} }
resource LocalResource { action read }
service LocalService { method ping {} }
event LocalEvent { payload { value: string } }
`
	const exportedSource = `
pub enum Mode { ACTIVE }
pub data Direct { id: string }
pub config SettingsConfig eternal { enabled: bool }
pub actor ClientActor { via client {} }
pub resource Record { action read }
pub service BackendService { method get { output Payload } }
ext service StorageService { method get { output Payload } }
pub event ChangedEvent { payload { value: Payload } }
ext event StoredEvent { payload { value: Payload } }
`
	for _, exported := range []bool{false, true} {
		name := "private only"
		if exported {
			name = "public contract"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "contract.skel")
			source := privateSource
			if exported {
				source += exportedSource
			}
			if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			parsed, err := compiler.Compile(compiler.Option{SkelIn: path})
			if err != nil {
				t.Fatal(err)
			}
			outputDir := filepath.Join(root, "generated")
			gen := newGen(Option{
				Domain: parsed.Domain, View: mustView(t, view.ModeRegular, parsed.Domain), Mode: view.ModeRegular,
				PackageName: "demo", PubImportPath: "example.com/demopub", Out: outputDir,
			})
			gen.genFacadeGo()
			if err := gen.Renderer.Err(); err != nil {
				t.Fatal(err)
			}
			if !exported {
				if _, err := os.Stat(filepath.Join(outputDir, facadeGoFilename)); !os.IsNotExist(err) {
					t.Fatalf("private-only domain generated a public facade: %v", err)
				}
				return
			}
			content := readFacadeGoForTest(t, outputDir)
			for _, name := range []string{
				"Mode", "Direct", "State", "Payload", "SettingsConfig", "ClientActor",
				"BackendServiceClient", "StorageServiceServer", "ChangedEventListener", "StoredEventEmitter",
			} {
				if !strings.Contains(content, "type "+name+" = demopub."+name) {
					t.Fatalf("public facade omitted %s: %s", name, content)
				}
			}
			if !strings.Contains(content, "RecordReadPermission") || strings.Contains(content, "Local") {
				t.Fatalf("facade differs from selected public contract: %s", content)
			}
		})
	}
}

func TestFacadeGoRendersActorAuthService(t *testing.T) {
	credential := &schema.Data{
		Name: "PublicActorCredential",
		Members: []*schema.DataMember{
			{Name: "token", Type: codegentest.StringType()},
		},
	}
	info := &schema.Data{
		Name: "PublicActorInfo",
		Members: []*schema.DataMember{
			{Name: "userId", Type: codegentest.StringType()},
		},
	}
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.auth",
		Actors: []*schema.Actor{
			{
				Pub:  true,
				Name: "PublicActor",
				Vias: []*schema.ActorVia{codegentest.ActorVia(schema.ActorViaClient)},
				Auth: new(schema.ActorAuth{Credential: credential, Info: info}),
			},
		},
	})
	outputDir := filepath.Join(t.TempDir(), "auth")
	gen := newGen(Option{
		Domain:        pkg,
		View:          mustView(t, view.ModeRegular, pkg),
		Mode:          view.ModeRegular,
		PackageName:   "auth",
		Out:           outputDir,
		PubImportPath: "github.com/acme/skel/demo/authpub",
	})

	gen.genFacadeGo()

	content := readFacadeGoForTest(t, outputDir)
	for _, expected := range []string{
		`import "github.com/acme/skel/demo/authpub"`,
		"type PublicActor = authpub.PublicActor",
		"type PublicActorCredential = authpub.PublicActorCredential",
		"type PublicActorInfo = authpub.PublicActorInfo",
		"type PublicActorAuthServiceServer = authpub.PublicActorAuthServiceServer",
		"type DefaultPublicActorAuthServiceServer = authpub.DefaultPublicActorAuthServiceServer",
		"type PublicActorAuthServiceServerER = authpub.PublicActorAuthServiceServerER",
		"type DefaultPublicActorAuthServiceServerER = authpub.DefaultPublicActorAuthServiceServerER",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected pub.go to contain %q, got:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "PublicActorAuthServiceClient") {
		t.Fatalf("did not expect auth service client facade, got:\n%s", content)
	}
}

func TestFacadeGoRendersResourcePermissions(t *testing.T) {
	pkg := buildSchemaDomainForTest(t, schema.DomainSpec{
		Name: "demo.app",
		Resources: []*schema.Resource{
			{
				Pub:  true,
				Name: "User",
				Actions: []*schema.ResourceAction{
					{Name: "read"},
					{Name: "update"},
					{Name: "manage"},
				},
			},
		},
	})
	outputDir := filepath.Join(t.TempDir(), "app")
	gen := newGen(Option{
		Domain:        pkg,
		View:          mustView(t, view.ModeRegular, pkg),
		Mode:          view.ModeRegular,
		PackageName:   "app",
		Out:           outputDir,
		PubImportPath: "github.com/acme/skel/demo/apppub",
	})

	gen.genFacadeGo()

	content := readFacadeGoForTest(t, outputDir)
	normalizedContent := strings.Join(strings.Fields(content), " ")
	for _, expected := range []string{
		"UserReadPermission = apppub.UserReadPermission",
		"UserUpdatePermission = apppub.UserUpdatePermission",
		"UserManagePermission = apppub.UserManagePermission",
	} {
		if !strings.Contains(normalizedContent, expected) {
			t.Fatalf("expected pub.go to contain %q, got:\n%s", expected, content)
		}
	}
}

func readFacadeGoForTest(t *testing.T, outputDir string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(outputDir, facadeGoFilename))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	return string(content)
}
