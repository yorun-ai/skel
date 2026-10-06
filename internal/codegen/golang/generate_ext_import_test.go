package golang_test

import (
	"path/filepath"
	"strings"
	"testing"

	"go.yorun.ai/skel"
	"go.yorun.ai/skel/internal/testutil"
)

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
	if _, err := skel.CompileSkeleton(skel.Input{SkelIn: sharedSource, Strict: true}, skel.SkeletonOption{Out: sharedSkel, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	sharedGo := filepath.Join(root, "contractspub")
	const sharedModule = "example.com/contractspub"
	if _, err := skel.CompileGolang(skel.Input{SkelIn: sharedSkel, Strict: true}, skel.GolangOption{CompilerVersion: "v0.0.0-dev", Out: sharedGo, Module: sharedModule, PubOnly: true, AsModule: true}); err != nil {
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
	if _, err := skel.CompileSkeleton(skel.Input{SkelIn: ownerSource, SkelImports: imports, Strict: true}, skel.SkeletonOption{Out: ownerSkel, PubOnly: true}); err != nil {
		t.Fatal(err)
	}
	input := skel.Input{SkelIn: ownerSkel, SkelImports: imports, Strict: true}
	parsed, err := skel.Parse(input)
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
	if _, err := skel.CompileGolang(input, skel.GolangOption{
		CompilerVersion: "v0.0.0-dev", Out: regular, Module: "example.com/audit",
		PubOut: pub, PubModule: "example.com/auditpub", AsModule: true,
		Imports: map[string]string{"common.contracts": sharedModule},
	}); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{pub, regular} {
		mod := readFileForTest(t, filepath.Join(directory, "go.mod"))
		if !strings.Contains(mod, sharedModule+" ") || !strings.Contains(mod, "go.yorun.ai/vine "+skel.DefaultGolangVineVersion) {
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
	testutil.Go(t, pub, "build", "-mod=mod", "./...")
	testutil.Go(t, regular, "test", "-mod=mod", "./...")
}
