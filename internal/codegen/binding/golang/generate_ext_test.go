package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
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
	if _, err := api.CompileSkeleton(api.Input{SkelIn: input, Strict: true}, api.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	pub := filepath.Join(root, "storagepub")
	regular := filepath.Join(root, "storage")
	if _, err := api.CompileGolang(api.Input{SkelIn: input, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/storage", PubOut: pub, PubModule: "example.com/storagepub", AsModule: true}); err != nil {
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
		if _, err := api.CompileGolang(api.Input{SkelIn: skelOut, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/standalone", AsModule: true, PubOnly: pubOnly}); err != nil {
			t.Fatal(err)
		}
		testutil.UseLocalSkel(t, out)
		testutil.Go(t, out, "build", "-mod=mod", "./...")
	}
	testutil.UseLocalSkel(t, pub)
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.Go(t, regular, "mod", "edit", "-replace=example.com/storagepub="+pub)
	testutil.UseLocalSkel(t, regular)
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
}

func TestExtensionEventSplitContracts(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	input := filepath.Join(root, "event.skel")
	writeFileForTest(t, input, `domain demo.audit
data Detail { message: string }
ext event AuditRecordedEvent { @sensitive payload { detail: Detail at: timestamp } }
pub event AuditStoredEvent { payload {} }
event PrivateEvent { payload {} }
`)
	skelOut := filepath.Join(root, "skel")
	if _, err := api.CompileSkeleton(api.Input{SkelIn: input, Strict: true}, api.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	pub := filepath.Join(root, "auditpub")
	regular := filepath.Join(root, "audit")
	if _, err := api.CompileGolang(api.Input{SkelIn: input, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/audit", PubOut: pub, PubModule: "example.com/auditpub", AsModule: true}); err != nil {
		t.Fatal(err)
	}
	read := func(directory string, name string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(content)
	}
	pubEvent := read(pub, "event.go")
	if !strings.Contains(pubEvent, "type AuditRecordedEventEmitter interface") || strings.Contains(pubEvent, "AuditRecordedEventListener") || strings.Contains(pubEvent, "PrivateEvent") {
		t.Fatalf("incorrect public contract: %s", pubEvent)
	}
	regularEvent := read(regular, "event.go")
	if !strings.Contains(regularEvent, "type AuditRecordedEventListener interface") || strings.Contains(regularEvent, "type AuditRecordedEventEmitter interface") || strings.Contains(regularEvent, "type AuditRecordedEvent struct") {
		t.Fatalf("incorrect regular contract: %s", regularEvent)
	}
	facade := read(regular, "pub.go")
	for _, expected := range []string{"type AuditRecordedEvent =", "type AuditRecordedEventEmitter =", "var NewAuditRecordedEventEmitter ="} {
		if !strings.Contains(facade, expected) {
			t.Fatalf("missing %s in facade: %s", expected, facade)
		}
	}
	for _, directory := range []string{pub, regular} {
		if !strings.Contains(read(directory, "schema.go"), "Ext: true") {
			t.Fatal("extension runtime flag missing")
		}
	}
	writeFileForTest(t, filepath.Join(regular, "ext_event_test.go"), `package audit_test
import (
 "reflect"
 "testing"
 audit "example.com/audit"
 pub "example.com/auditpub"
 "go.yorun.ai/vine/core/ex"
 "go.yorun.ai/vine/core/skel"
)
type listener struct { audit.DefaultAuditRecordedEventListener; message string }
func (l *listener) OnAuditRecorded(event *pub.AuditRecordedEvent) { l.message = event.Detail.Message }
type errorListener struct { audit.DefaultAuditRecordedEventListenerER }
func (*errorListener) OnAuditRecorded(event *pub.AuditRecordedEvent) ex.Error { return nil }
var _ audit.AuditRecordedEventListener = (*listener)(nil)
var _ audit.AuditRecordedEventListenerER = (*errorListener)(nil)
var _ audit.AuditRecordedEventEmitter = (pub.AuditRecordedEventEmitter)(nil)
func TestExtensionContract(t *testing.T) {
 if reflect.TypeFor[audit.AuditRecordedEvent]() != reflect.TypeFor[pub.AuditRecordedEvent]() { t.Fatal("payload types differ") }
 handler := &listener{}
 handler.OnAuditRecorded(&pub.AuditRecordedEvent{Detail: pub.Detail{Message:"received"}})
 if handler.message != "received" { t.Fatal("wrong payload") }
 for _, domain := range skel.RegisteredDomainSchemas() {
  if domain.Domain != "demo.audit" { continue }
  for _, event := range domain.Events {
   if event.SkelName == "demo.audit.AuditRecordedEvent" {
    if !event.Ext || !event.Pub || !event.Sensitive { t.Fatalf("wrong schema: %+v", event) }
    return
   }
  }
 }
 t.Fatal("extension event schema not registered")
}
`)
	testutil.Go(t, regular, "mod", "edit", "-replace=example.com/auditpub="+pub)
	testutil.UseLocalSkel(t, pub)
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.UseLocalSkel(t, regular)
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
	for _, pubOnly := range []bool{false, true} {
		out := filepath.Join(t.TempDir(), "standalone")
		if _, err := api.CompileGolang(api.Input{SkelIn: skelOut, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/auditpub", AsModule: true, PubOnly: pubOnly}); err != nil {
			t.Fatal(err)
		}
		code := read(out, "event.go")
		if strings.Contains(code, "PrivateEvent") || !strings.Contains(code, "AuditRecordedEventEmitter") || (pubOnly && strings.Contains(code, "AuditRecordedEventListener")) || (!pubOnly && !strings.Contains(code, "AuditRecordedEventListener")) {
			t.Fatalf("incorrect standalone event: %s", code)
		}
		testutil.UseLocalSkel(t, out)
		testutil.Go(t, out, "build", "-mod=mod", "./...")
	}
}

func TestExtensionContractsImportExternalGenericData(t *testing.T) {
	testutil.RequireToolchain(t)
	root := t.TempDir()
	sharedSource := filepath.Join(root, "shared.skel")
	writeFileForTest(t, sharedSource, `domain common.contracts
pub data Page<TItem> { items: list<TItem> }
pub data Record { id: string }
data Hidden { value: string }
`)
	sharedSkel := filepath.Join(root, "shared-skel")
	if _, err := api.CompileSkeleton(api.Input{SkelIn: sharedSource, Strict: true}, api.SkeletonOption{Out: sharedSkel, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	sharedGo := filepath.Join(root, "contractspub")
	const sharedModule = "example.com/contractspub"
	if _, err := api.CompileGolang(api.Input{SkelIn: sharedSkel, Strict: true}, api.GolangOption{CompilerVersion: "v0.0.0-dev", Out: sharedGo, Module: sharedModule, PubOnly: true, AsModule: true}); err != nil {
		t.Fatal(err)
	}
	ownerSource := filepath.Join(root, "audit.skel")
	writeFileForTest(t, ownerSource, `domain demo.audit
import common.contracts as shared
data Envelope { records: shared.Page<shared.Record> }
data PrivateData { value: string }
ext service AuditService {
    method record { input { envelope: Envelope } output shared.Page<shared.Record> }
}
ext event AuditRecordedEvent { payload { envelope: Envelope records: shared.Page<shared.Record> } }
`)
	imports := map[string]string{"common.contracts": sharedSkel}
	ownerSkel := filepath.Join(root, "audit-skel")
	if _, err := api.CompileSkeleton(api.Input{SkelIn: ownerSource, SkelImports: imports, Strict: true}, api.SkeletonOption{Out: ownerSkel, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	input := api.Input{SkelIn: ownerSkel, SkelImports: imports, Strict: true}
	parsed, err := api.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	data := parsed.Domain.Data()
	if len(data) != 1 || data[0].Name != "Envelope" {
		t.Fatalf("public dependency closure must contain only the required local data: %+v", data)
	}
	if len(parsed.Domain.Services()) != 1 || !parsed.Domain.Services()[0].Ext || len(parsed.Domain.Events()) != 1 || !parsed.Domain.Events()[0].Ext {
		t.Fatal("public Skel round trip lost the extension contracts")
	}
	kind := data[0].Members[0].Type
	if kind.ExternalDomain != "common.contracts" || kind.ExternalAlias != "shared" || kind.Data.Name != "Page" || len(kind.TypeArguments) != 1 || kind.TypeArguments[0].Data.Name != "Record" {
		t.Fatalf("public Skel round trip lost the external generic or alias: %+v", kind)
	}
	pub := filepath.Join(root, "auditpub")
	regular := filepath.Join(root, "audit")
	if _, err := api.CompileGolang(input, api.GolangOption{
		CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/audit",
		PubOut: pub, PubModule: "example.com/auditpub", AsModule: true,
		Imports: map[string]string{"common.contracts": sharedModule},
	}); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{pub, regular} {
		mod := readFileForTest(t, filepath.Join(directory, "go.mod"))
		if !strings.Contains(mod, sharedModule+" ") || !strings.Contains(mod, "go.yorun.ai/vine "+api.DefaultGolangVineVersion) {
			t.Fatalf("generated module lost its contract or runtime dependency: %s", mod)
		}
		testutil.Go(t, directory, "mod", "edit", "-replace="+sharedModule+"="+sharedGo)
	}
	testutil.Go(t, regular, "mod", "edit", "-replace=example.com/auditpub="+pub)
	writeFileForTest(t, filepath.Join(regular, "ext_import_test.go"), `package audit_test
import (
    "reflect"
    "testing"
    audit "example.com/audit"
    pub "example.com/auditpub"
    shared "example.com/contractspub"
)
type implementation struct { pub.DefaultAuditServiceServer }
func (*implementation) Record(envelope pub.Envelope) shared.Page[shared.Record] { return envelope.Records }
type listener struct { audit.DefaultAuditRecordedEventListener; records shared.Page[shared.Record] }
func (l *listener) OnAuditRecorded(event *pub.AuditRecordedEvent) { l.records = event.Records }
var _ pub.AuditServiceServer = (*implementation)(nil)
var _ audit.AuditServiceServer = (*implementation)(nil)
var _ audit.AuditRecordedEventListener = (*listener)(nil)
func TestImportedGenericContract(t *testing.T) {
    records := shared.Page[shared.Record]{Items: []shared.Record{{Id:"audit"}}}
    envelope := audit.Envelope{Records: records}
    var server audit.AuditServiceServer = &implementation{}
    if server.Record(envelope).Items[0].Id != "audit" { t.Fatal("wrong imported service result") }
    handler := &listener{}
    handler.OnAuditRecorded(&audit.AuditRecordedEvent{Envelope: envelope, Records: records})
    if handler.records.Items[0].Id != "audit" { t.Fatal("wrong imported event payload") }
    if reflect.TypeFor[audit.AuditRecordedEvent]() != reflect.TypeFor[pub.AuditRecordedEvent]() { t.Fatal("event payload types differ") }
    if reflect.TypeFor[audit.AuditRecordedEventEmitter]() != reflect.TypeFor[pub.AuditRecordedEventEmitter]() { t.Fatal("emitter types differ") }
}
`)
	testutil.UseLocalSkel(t, pub)
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.UseLocalSkel(t, regular)
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
}
