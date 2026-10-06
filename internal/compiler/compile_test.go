package compiler

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestCompileDirectory(t *testing.T) {
	skelDir := t.TempDir()
	writeFile(t, filepath.Join(skelDir, "domain.skel"), describedUserDomain)
	writeFile(t, filepath.Join(skelDir, "user.skel"), `
domain demo.user

actor ClientActor { via client {} }

api service UserApiService { auth required
    for ClientActor

    method getUser {
        output User
    }
}

data User {
    id: string
}

config SiteConfig eternal {
    title: string
}
`)

	result, err := Compile(Option{SkelIn: skelDir})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if result.Domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain name: %s", result.Domain.Name())
	}
	if result.Domain.Description() != "User domain" {
		t.Fatalf("unexpected domain description: %s", result.Domain.Description())
	}
	if len(result.Domain.Data()) != 1 || result.Domain.Data()[0].Name != "User" {
		t.Fatalf("unexpected data: %#v", result.Domain.Data())
	}
	if len(result.Domain.Configs()) != 1 || result.Domain.Configs()[0].Name != "SiteConfig" {
		t.Fatalf("unexpected configs: %#v", result.Domain.Configs())
	}
	if len(result.Domain.Services()) != 1 || result.Domain.Services()[0].Name != "UserApiService" {
		t.Fatalf("unexpected services: %#v", result.Domain.Services())
	}
}

func TestCompileRequiresTransitiveImports(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	userDir := filepath.Join(root, "user")
	bookerDir := filepath.Join(root, "booker")
	mustMkdirAll(t, appDir, userDir, bookerDir)

	writeFile(t, filepath.Join(appDir, "domain.skel"), "domain app\n")
	writeFile(t, filepath.Join(appDir, "actor.skel"), `
domain app

pub actor UserActor {
    via client {}
}
`)
	writeFile(t, filepath.Join(userDir, "domain.skel"), "domain user\n")
	writeFile(t, filepath.Join(userDir, "service.skel"), `
domain user

import app

api service UserApiService { auth required
    for app.UserActor

    method getUser {
        output UserSummary
    }
}

pub data UserSummary {
    id: string
}
`)
	writeFile(t, filepath.Join(bookerDir, "domain.skel"), "domain booker\n")
	writeFile(t, filepath.Join(bookerDir, "types.skel"), `
domain booker

import user

pub data Loan {
    borrower: user.UserSummary
}
`)

	_, err := Compile(Option{
		SkelIn:      bookerDir,
		SkelImports: map[string]string{"user": userDir},
	})
	if err == nil || !strings.Contains(err.Error(), "skel import app not found") {
		t.Fatalf("expected missing transitive import error, got %v", err)
	}

	result, err := Compile(Option{
		SkelIn: bookerDir,
		SkelImports: map[string]string{
			"app":  appDir,
			"user": userDir,
		},
	})
	if err != nil {
		t.Fatalf("Compile() with transitive imports error = %v", err)
	}

	if len(result.Domain.Data()) != 1 || result.Domain.Data()[0].Name != "Loan" {
		t.Fatalf("unexpected data: %#v", result.Domain.Data())
	}
	if len(result.Domain.Imports()) != 1 || result.Domain.Imports()[0].Name != "user" {
		t.Fatalf("unexpected imports: %#v", result.Domain.Imports())
	}
	userDomain := result.Domain.Imports()[0].Domain
	if len(userDomain.Imports()) != 1 || userDomain.Imports()[0].Name != "app" {
		t.Fatalf("unexpected transitive imports: %#v", userDomain.Imports())
	}
}

func TestCompileNormalizesImportedDomainLocalTypeReferences(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")
	appDir := filepath.Join(root, "app")
	mustMkdirAll(t, baseDir, appDir)

	writeFile(t, filepath.Join(baseDir, "domain.skel"), "domain base\n")
	writeFile(t, filepath.Join(baseDir, "types.skel"), `
domain base

pub enum ItemType {
    STANDARD
}

pub data Detail {
    name: string
}

pub data Box<TItem> {
    value: TItem
}

pub data Item {
    type: ItemType
    detail: Detail
    boxedType: Box<ItemType>
    types: list<ItemType>
    detailsByType: map<ItemType, Detail>
    boxedDetails: Box<list<Detail>>
}

pub actor BaseActor {
    via client {}
    auth {
        credential {
            subject: string
        }
        info {
            type: ItemType
            detail: Detail
        }
    }
}
`)
	writeFile(t, filepath.Join(appDir, "domain.skel"), "domain app\n")
	writeFile(t, filepath.Join(appDir, "types.skel"), `
domain app

import base

data AppItem {
    item: base.Item
}
`)

	result, err := Compile(Option{
		SkelIn:      appDir,
		SkelImports: map[string]string{"base": baseDir},
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	baseDomain := result.Domain.Imports()[0].Domain
	var box, item *schema.Data
	for _, dataType := range baseDomain.Data() {
		switch dataType.Name {
		case "Box":
			box = dataType
		case "Item":
			item = dataType
		}
	}
	if box == nil || item == nil {
		t.Fatalf("expected imported Box and Item data: %+v", baseDomain.Data())
	}
	if got := box.Members[0].Type.Kind; got != schema.TypeKindTypeParameter {
		t.Fatalf("generic member kind = %d, want %d", got, schema.TypeKindTypeParameter)
	}
	members := make(map[string]*schema.Type, len(item.Members))
	for _, member := range item.Members {
		members[member.Name] = member.Type
	}
	if got := members["type"].Kind; got != schema.TypeKindEnum {
		t.Fatalf("enum member kind = %d, want %d", got, schema.TypeKindEnum)
	}
	if got := members["detail"].Kind; got != schema.TypeKindData {
		t.Fatalf("data member kind = %d, want %d", got, schema.TypeKindData)
	}
	if got := members["boxedType"].TypeArguments[0].Kind; got != schema.TypeKindEnum {
		t.Fatalf("generic type argument kind = %d, want %d", got, schema.TypeKindEnum)
	}
	if got := members["types"].List.Element.Kind; got != schema.TypeKindEnum {
		t.Fatalf("list element kind = %d, want %d", got, schema.TypeKindEnum)
	}
	if got := members["detailsByType"].Map.Key.Kind; got != schema.TypeKindEnum {
		t.Fatalf("map key kind = %d, want %d", got, schema.TypeKindEnum)
	}
	if got := members["detailsByType"].Map.Value.Kind; got != schema.TypeKindData {
		t.Fatalf("map value kind = %d, want %d", got, schema.TypeKindData)
	}
	if got := members["boxedDetails"].TypeArguments[0].List.Element.Kind; got != schema.TypeKindData {
		t.Fatalf("nested generic list element kind = %d, want %d", got, schema.TypeKindData)
	}
	authInfo := baseDomain.Actors()[0].Auth.Info
	wantKinds := []schema.TypeKind{schema.TypeKindEnum, schema.TypeKindData}
	for index, member := range authInfo.Members {
		if member.Type.Kind != wantKinds[index] {
			t.Fatalf("actor info member %s kind = %d, want %d", member.Name, member.Type.Kind, wantKinds[index])
		}
	}
}

func TestCompileNormalizesTransitiveImportedTypeReferences(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")
	userDir := filepath.Join(root, "user")
	appDir := filepath.Join(root, "app")
	mustMkdirAll(t, baseDir, userDir, appDir)

	writeFile(t, filepath.Join(baseDir, "domain.skel"), "domain base\n")
	writeFile(t, filepath.Join(baseDir, "types.skel"), `
domain base

pub enum UserStatus {
    ACTIVE
}
`)
	writeFile(t, filepath.Join(userDir, "domain.skel"), "domain user\n")
	writeFile(t, filepath.Join(userDir, "types.skel"), `
domain user

import base

pub data User {
    status: base.UserStatus
}
`)
	writeFile(t, filepath.Join(appDir, "domain.skel"), "domain app\n")
	writeFile(t, filepath.Join(appDir, "types.skel"), `
domain app

import user

data AppUser {
    user: user.User
}
`)

	result, err := Compile(Option{
		SkelIn: appDir,
		SkelImports: map[string]string{
			"base": baseDir,
			"user": userDir,
		},
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	userDomain := result.Domain.Imports()[0].Domain
	if got := userDomain.Data()[0].Members[0].Type.Kind; got != schema.TypeKindEnum {
		t.Fatalf("transitive imported enum kind = %d, want %d", got, schema.TypeKindEnum)
	}
}

func TestCompileRejectsCyclicTransitiveImports(t *testing.T) {
	root := t.TempDir()
	firstDir := filepath.Join(root, "first")
	secondDir := filepath.Join(root, "second")
	targetDir := filepath.Join(root, "target")
	mustMkdirAll(t, firstDir, secondDir, targetDir)

	writeFile(t, filepath.Join(firstDir, "domain.skel"), "domain first\n")
	writeFile(t, filepath.Join(firstDir, "types.skel"), "domain first\nimport second\n")
	writeFile(t, filepath.Join(secondDir, "domain.skel"), "domain second\n")
	writeFile(t, filepath.Join(secondDir, "types.skel"), "domain second\nimport first\n")
	writeFile(t, filepath.Join(targetDir, "domain.skel"), "domain target\n")
	writeFile(t, filepath.Join(targetDir, "types.skel"), "domain target\nimport first\n")

	_, err := Compile(Option{
		SkelIn: targetDir,
		SkelImports: map[string]string{
			"first":  firstDir,
			"second": secondDir,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "cyclic skel import involving") {
		t.Fatalf("expected cyclic import error, got %v", err)
	}
}

func TestCompileRejectsTypeParameterOutsideDeclaringData(t *testing.T) {
	tests := []struct {
		name        string
		declaration string
	}{
		{
			name: "actor auth info",
			declaration: `
actor DemoActor {
    via client {}
    auth {
        credential {
            subject: string
        }
        info {
            value: TItem
        }
    }
}
`,
		},
		{
			name: "resource check",
			declaration: `
resource DemoResource {
    check byValue {
        input {
            value: TItem
        }
    }
    action read
}
`,
		},
		{
			name: "service method",
			declaration: `
pub service DemoService {
    method getValue {
        output TItem
    }
}
`,
		},
		{
			name: "task trigger",
			declaration: `
task DemoTask {
    trigger manually {
        input {
            value: TItem
        }
    }
}
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			skelFile := filepath.Join(t.TempDir(), "types.skel")
			writeFile(t, skelFile, `
domain demo.types

data Wrapper<TItem> {
    value: TItem
}

`+test.declaration)

			_, err := Compile(Option{SkelIn: skelFile})
			if err == nil || !strings.Contains(err.Error(), "definition of TItem not found") {
				t.Fatalf("expected out-of-scope type parameter error, got %v", err)
			}
		})
	}
}

func TestCompileRejectsConfigReferenceFromActorAuthInfo(t *testing.T) {
	skelFile := filepath.Join(t.TempDir(), "types.skel")
	writeFile(t, skelFile, `
domain demo.types

config SessionConfig eternal {
    issuer: string
}

actor DemoActor {
    via client {}
    auth {
        credential {
            subject: string
        }
        info {
            session: SessionConfig
        }
    }
}
`)

	tests := []struct {
		name  string
		parse func() error
	}{
		{
			name: "primary domain",
			parse: func() error {
				_, err := Compile(Option{SkelIn: skelFile})
				return err
			},
		},
		{
			name: "imported domain",
			parse: func() error {
				_, err := CompileImport(Option{SkelIn: skelFile})
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.parse()
			if err == nil || !strings.Contains(err.Error(), "config SessionConfig cannot be used as a value type") {
				t.Fatalf("expected actor info config reference error, got %v", err)
			}
		})
	}
}

func TestCompileImportDoesNotRequireImportedDomains(t *testing.T) {
	skelDir := t.TempDir()
	writeFile(t, filepath.Join(skelDir, "domain.skel"), "domain demo.booker\n")
	writeFile(t, filepath.Join(skelDir, "types.skel"), `
domain demo.booker

import demo.user as user

data Booking {
    owner: user.User
}
`)

	result, err := CompileImport(Option{SkelIn: skelDir})
	if err != nil {
		t.Fatalf("CompileImport() error = %v", err)
	}

	if result.Domain.Name() != "demo.booker" {
		t.Fatalf("unexpected domain name: %s", result.Domain.Name())
	}
	if result.Domain.ReferenceName("user.User") != "demo.user.User" {
		t.Fatalf("unexpected import aliases: %#v", result.Domain.Imports())
	}
}

func TestCompileReturnsErrorWhenDomainFileMissing(t *testing.T) {
	skelDir := t.TempDir()
	writeFile(t, filepath.Join(skelDir, "user.skel"), "data User { id: string }\n")

	_, err := Compile(Option{SkelIn: skelDir})
	expectErrorContains(t, err, "domain.skel not found")
}

func TestCompileSingleSkelFile(t *testing.T) {
	skelFile := filepath.Join(t.TempDir(), "user.skel")
	writeFile(t, skelFile, `@desc("User domain")
domain demo.user

data User {
    id: string
}
`)

	result, err := Compile(Option{SkelIn: skelFile})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	if result.Domain.Name() != "demo.user" || len(result.Domain.Data()) != 1 {
		t.Fatalf("unexpected domain: %#v", result.Domain)
	}
}

func TestCompileReturnsSemanticErrorForMissingDefinition(t *testing.T) {
	skelFile := filepath.Join(t.TempDir(), "user.skel")
	writeFile(t, skelFile, `domain demo.user

actor ClientActor { via client {} }

api service UserApiService { auth required
    for ClientActor

    method getUser {
        output MissingUser
    }
}
`)

	_, err := Compile(Option{SkelIn: skelFile})
	if err == nil || !strings.Contains(err.Error(), "definition of MissingUser not found") {
		t.Fatalf("expected semantic error, got %v", err)
	}
}

func TestImportQualifiers(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, "shared.skel")
	writeFile(t, shared, `domain ws.sandbox
pub actor SandboxActor { via client {} }
pub data Box<TValue> { value: TValue }
pub enum State { READY }
pub resource Document { action read }
`)
	other := filepath.Join(root, "other.skel")
	writeFile(t, other, "domain other.sandbox\npub data Item {}\n")
	var webHash, serviceHash string
	for _, test := range []struct {
		name, declaration, qualifier string
		valid                        bool
	}{
		{"full domain", "ws.sandbox", "ws.sandbox", true},
		{"explicit alias", "ws.sandbox as sandbox", "sandbox", true},
		{"custom alias", "ws.sandbox as sb", "sb", true},
		{"no implicit alias", "ws.sandbox", "sandbox", false},
		{"alias replaces domain", "ws.sandbox as sb", "ws.sandbox", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := filepath.Join(t.TempDir(), "domain.skel")
			source := `domain demo.proxy
import IMPORT
import other.sandbox
pub data Payload { value: REF.Box<REF.State> other: other.sandbox.Item }
web DeepseekWeb { auth required  for REF.SandboxActor }
api service ProxyApiService { auth required  for REF.SandboxActor method ping { require REF.Document:read } }
`
			source = strings.ReplaceAll(strings.ReplaceAll(source, "IMPORT", test.declaration), "REF", test.qualifier)
			writeFile(t, input, source)
			result, err := Compile(Option{SkelIn: input, SkelImports: map[string]string{"ws.sandbox": shared, "other.sandbox": other}})
			if !test.valid {
				if err == nil {
					t.Fatal("expected unresolved reference")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Domain.ReferenceName(test.qualifier+".User") != "ws.sandbox.User" {
				t.Fatalf("unexpected imports: %v", result.Domain.Imports())
			}
			if webHash == "" {
				webHash = result.Domain.Webs()[0].Hash
				serviceHash = result.Domain.Services()[0].Hash
			}
			if result.Domain.Webs()[0].Hash != webHash || result.Domain.Services()[0].Hash != serviceHash {
				t.Fatal("import spelling changed compatibility hashes")
			}
			member := result.Domain.Data()[0].Members[0].Type
			if member.SkelName != "ws.sandbox.Box" || member.TypeArguments[0].SkelName != "ws.sandbox.State" {
				t.Fatalf("incorrect canonical type: %+v", member)
			}
		})
	}
}
