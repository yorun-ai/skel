package golang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel"
	"go.yorun.ai/skel/internal/testutil"
)

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
	if _, err := skel.CompileSkeleton(skel.Input{SkelIn: input, Strict: true}, skel.SkeletonOption{Out: skelOut, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	pub := filepath.Join(root, "auditpub")
	regular := filepath.Join(root, "audit")
	if _, err := skel.CompileGolang(skel.Input{SkelIn: input, Strict: true}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/audit", PubOut: pub, PubModule: "example.com/auditpub", AsModule: true}); err != nil {
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
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
	for _, pubOnly := range []bool{false, true} {
		out := filepath.Join(t.TempDir(), "standalone")
		if _, err := skel.CompileGolang(skel.Input{SkelIn: skelOut, Strict: true}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: out, Module: "example.com/auditpub", AsModule: true, PubOnly: pubOnly}); err != nil {
			t.Fatal(err)
		}
		code := read(out, "event.go")
		if strings.Contains(code, "PrivateEvent") || !strings.Contains(code, "AuditRecordedEventEmitter") || (pubOnly && strings.Contains(code, "AuditRecordedEventListener")) || (!pubOnly && !strings.Contains(code, "AuditRecordedEventListener")) {
			t.Fatalf("incorrect standalone event: %s", code)
		}
		testutil.Go(t, out, "build", "-mod=mod", "./...")
	}
}
