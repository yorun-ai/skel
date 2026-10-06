package schema_test

import (
	"bytes"
	"reflect"
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestSnapshotCodecRoundTrip(t *testing.T) {
	document := new(schema.Document{
		Format:        schema.Format,
		FormatVersion: schema.FormatVersion,
		Domain:        "demo.user",
		Declarations: []*schema.Declaration{new(schema.Declaration{
			Pub:      true,
			Name:     "User",
			Kind:     schema.DeclarationTypeData,
			SkelName: "demo.user.User",
			Data:     new(schema.DataSchema{Members: []*schema.Member{}}),
		})},
	})

	if err := schema.Validate(document); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	var encoded bytes.Buffer
	if err := schema.Encode(&encoded, document); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	decoded, err := schema.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, document) {
		t.Fatalf("round trip mismatch:\nwant: %#v\ngot:  %#v", document, decoded)
	}
}

func TestExtensionServiceCodec(t *testing.T) {
	document := new(schema.Document{
		Format: schema.Format, FormatVersion: schema.FormatVersion, Domain: "demo.storage",
		Declarations: []*schema.Declaration{new(schema.Declaration{
			Pub: true, Name: "StorageService", Kind: schema.DeclarationTypeService, SkelName: "demo.storage.StorageService",
			Service: new(schema.ServiceSchema{Ext: true, Auth: schema.AuthModeUnset, Audiences: []*schema.Audience{}, Methods: []*schema.Method{}}),
		})},
	})
	var encoded bytes.Buffer
	if err := schema.Encode(&encoded, document); err != nil {
		t.Fatal(err)
	}
	decoded, err := schema.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil || !decoded.Declarations[0].Service.Ext {
		t.Fatalf("lost extension contract: %v", err)
	}
	legacy := bytes.ReplaceAll(encoded.Bytes(), []byte(`"ext"`), []byte(`"open"`))
	if _, err := schema.Decode(bytes.NewReader(legacy)); err == nil {
		t.Fatal("accepted removed open schema field")
	}
	document.Declarations[0].Pub = false
	if err := schema.Validate(document); err == nil {
		t.Fatal("accepted a non-public extension schema")
	}
	document.Declarations[0].Pub = true
	document.Declarations[0].Service.Api = true
	if err := schema.Validate(document); err == nil {
		t.Fatal("accepted extension API schema")
	}
}

func TestExtensionEventCodec(t *testing.T) {
	document := new(schema.Document{Format: schema.Format, FormatVersion: schema.FormatVersion, Domain: "demo.audit", Declarations: []*schema.Declaration{{Pub: true, Name: "AuditRecordedEvent", Kind: schema.DeclarationTypeEvent, SkelName: "demo.audit.AuditRecordedEvent", Data: &schema.DataSchema{Ext: true, Members: []*schema.Member{}}}}})
	var encoded bytes.Buffer
	if err := schema.Encode(&encoded, document); err != nil {
		t.Fatal(err)
	}
	decoded, err := schema.Decode(bytes.NewReader(encoded.Bytes()))
	if err != nil || !decoded.Declarations[0].Data.Ext {
		t.Fatalf("lost extension event: %v", err)
	}
	document.Declarations[0].Pub = false
	if err := schema.Validate(document); err == nil {
		t.Fatal("accepted non-public extension event")
	}
	document.Declarations[0].Pub = true
	for _, kind := range []schema.DeclarationType{schema.DeclarationTypeData, schema.DeclarationTypeConfig} {
		document.Declarations[0].Kind = kind
		if err := schema.Validate(document); err == nil {
			t.Fatalf("accepted ext %s", kind)
		}
	}
}
