package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel"
	"go.yorun.ai/skel/internal/testutil"
)

func TestExtServicePublicServer(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "service.skel")
	source := `domain demo.storage
 data Item { value: string }
 ext service StorageService {
   method get { input { key: string } output Item }
 }
 pub service LookupService { method ping {} }
 `
	writeFileForTest(t, input, source)
	skelOut := filepath.Join(root, "skel")
	if _, err := skel.CompileSkeleton(skel.Input{SkelIn: input, Strict: true}, skel.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	pub := filepath.Join(root, "storagepub")
	regular := filepath.Join(root, "storage")
	if _, err := skel.CompileGolang(skel.Input{SkelIn: input, Strict: true}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/storage", PubOut: pub, PubModule: "example.com/storagepub", AsModule: true}); err != nil {
		t.Fatal(err)
	}
	service, err := os.ReadFile(filepath.Join(pub, "service.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(service), "type StorageServiceServer interface") || strings.Contains(string(service), "type LookupServiceServer interface") {
		t.Fatalf("wrong public servers: %s", service)
	}

	if strings.Contains(string(service), "StorageServiceClient") || !strings.Contains(string(service), "rpc.ServiceSpecTypeServer") {
		t.Fatalf("extension public package must expose only the server: %s", service)
	}
	regularService, err := os.ReadFile(filepath.Join(regular, "service.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(regularService), "type StorageServiceClient interface") || strings.Contains(string(regularService), "type StorageServiceServer interface") || !strings.Contains(string(regularService), "rpc.ServiceSpecTypeClient") {
		t.Fatalf("extension regular package must generate the caller: %s", regularService)
	}
	facade, err := os.ReadFile(filepath.Join(regular, "pub.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(facade), "type StorageServiceServer =") || strings.Contains(string(facade), "type StorageServiceClient =") {
		t.Fatalf("incorrect extension facade: %s", facade)
	}
	testSource := `package storage_test
import (
 "reflect"
 "testing"
 storage "example.com/storage"
 pub "example.com/storagepub"
 "go.yorun.ai/vine/core/ex"
 "go.yorun.ai/vine/core/skel"
)
type implementation struct { pub.DefaultStorageServiceServer }
func (*implementation) Get(key string) pub.Item { return pub.Item{Value:key} }
type errorImplementation struct { pub.DefaultStorageServiceServerER }
func (*errorImplementation) Get(key string) (pub.Item, ex.Error) { return pub.Item{Value:key}, nil }
var _ pub.StorageServiceServer = (*implementation)(nil)
var _ storage.StorageServiceServer = (*implementation)(nil)
var _ pub.StorageServiceServerER = (*errorImplementation)(nil)
var _ storage.StorageServiceServerER = (*errorImplementation)(nil)
func TestPublicServer(t *testing.T) {
 if reflect.TypeFor[storage.StorageServiceServer]() != reflect.TypeFor[pub.StorageServiceServer]() { t.Fatal("server types differ") }
 var server storage.StorageServiceServer = &implementation{}
 if server.Get("ok").Value != "ok" { t.Fatal("wrong result") }
 for _, domain := range skel.RegisteredDomainSchemas() {
  if domain.Domain != "demo.storage" { continue }
  for _, service := range domain.Services {
   if service.SkelName == "demo.storage.StorageService" {
    if !service.Ext || service.ClientApi() { t.Fatalf("wrong runtime extension metadata: %+v", service) }
    return
   }
  }
 }
 t.Fatal("extension runtime schema not registered")
}
`
	writeFileForTest(t, filepath.Join(regular, "ext_test.go"), testSource)

	for _, pubOnly := range []bool{false, true} {
		out := filepath.Join(root, "full")
		if pubOnly {
			out = filepath.Join(root, "public-only")
		}
		if _, err := skel.CompileGolang(skel.Input{SkelIn: skelOut, Strict: true}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/standalone", AsModule: true, PubOnly: pubOnly}); err != nil {
			t.Fatal(err)
		}
		testutil.Go(t, out, "build", "-mod=mod", "./...")
	}
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.Go(t, regular, "mod", "edit", "-replace=example.com/storagepub="+pub)
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
}

func TestExtensionServicesAreExcludedFromApiClients(t *testing.T) {
	source := filepath.Join(t.TempDir(), "service.skel")
	// Legacy admission rules must not turn an extension into a client API.
	writeFileForTest(t, source, `domain demo.storage
actor ClientActor { via client {} }
ext service StorageService { noauth method get { output string } }
api service HealthApiService { for ClientActor via client noauth method ping {} }
`)
	for _, target := range []string{"go", "ts"} {
		t.Run(target, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "api")
			var err error
			if target == "go" {
				_, err = skel.CompileGolang(skel.Input{SkelIn: source}, skel.GolangOption{ApiOnly: true, Out: out})
			} else {
				_, err = skel.CompileTypeScript(skel.Input{SkelIn: source}, skel.TypeScriptOption{ApiOnly: true, Out: out})
			}
			if err != nil {
				t.Fatal(err)
			}
			contents, err := os.ReadFile(filepath.Join(out, "service."+target))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(contents), "StorageService") || !strings.Contains(string(contents), "HealthApiService") {
				t.Fatalf("incorrect API boundary: %s", contents)
			}
		})
	}
}
