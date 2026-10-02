package schema

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skelc/internal/model"
)

func TestEncodeDecodeRoundTripOmitsSourcePositions(t *testing.T) {
	document := newTestDocument(
		&Declaration{
			Pub: true, Name: "User", Kind: DeclarationTypeData, SkelName: "demo.user.User",
			Pos: model.Position{File: "/workspace/user.skel", Line: 2, Column: 10},
			Data: &DataSchema{Members: []*Member{{
				Name: "id", Type: scalarType("string"),
				Pos: model.Position{File: "/workspace/user.skel", Line: 3, Column: 5},
			}}},
		},
	)
	var output bytes.Buffer
	if err := Encode(&output, document); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "/workspace") {
		t.Fatalf("encoded schema contains source position:\n%s", output.String())
	}
	decoded, err := Decode(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Domain != document.Domain || len(decoded.Declarations) != 1 {
		t.Fatalf("unexpected decoded schema: %+v", decoded)
	}
	if decoded.Declarations[0].Pos != (model.Position{}) {
		t.Fatalf("decoded source position should be empty: %+v", decoded.Declarations[0].Pos)
	}
}

func TestDecodeRejectsUnknownFieldsAndTrailingValues(t *testing.T) {
	for _, input := range []string{
		`{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[],"unknown":true}`,
		`{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[]} {}`,
		`{"format":"yorun.skel.schema","formatVersion":1,"domain":"demo","declarations":[{"name":"TextConfig","type":"config","skelName":"demo.TextConfig","data":{"lifecycle":"eternal","members":[{"name":"text","unknownMemberField":true,"type":{"kind":"scalar","name":"string"}}]}}]}`,
	} {
		if _, err := Decode(strings.NewReader(input)); err == nil {
			t.Fatalf("expected strict decode to fail: %s", input)
		}
	}
}

func TestCompleteDocumentJSONGolden(t *testing.T) {
	document := completeTestDocument()
	var encoded bytes.Buffer
	if err := Encode(&encoded, document); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/complete.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded.Bytes(), want) {
		t.Fatalf("complete schema JSON changed (-want +got):\nwant:\n%s\ngot:\n%s", want, encoded.Bytes())
	}
	decoded, err := Decode(bytes.NewReader(want))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, document) {
		t.Fatalf("decoded complete schema differs:\nwant: %#v\ngot:  %#v", document, decoded)
	}
}

func completeTestDocument() *Document {
	stringType := func() *Type { return &Type{Kind: TypeKindScalar, Name: "string"} }
	intType := func() *Type { return &Type{Kind: TypeKindScalar, Name: "int"} }
	metadata := Metadata{
		Description: "Documented schema element.", Deprecated: true, DeprecatedReason: "Use the replacement.",
	}
	argument := func(name string, valueType *Type) *Argument {
		return &Argument{Metadata: metadata, Name: name, Example: `"example"`, Sensitive: true, Type: valueType}
	}

	return &Document{
		Format: Format, FormatVersion: FormatVersion, Domain: "demo.contract", Description: "Complete schema fixture.",
		Declarations: []*Declaration{
			{
				Metadata: metadata, Pub: true, Name: "Caller", Kind: DeclarationTypeActor, SkelName: "demo.contract.Caller",
				Actor: &ActorSchema{
					Vias: []*ActorVia{{Name: "http"}}, AuthEnabled: true,
					AuthCredential: &DataSchema{Members: []*Member{{Metadata: metadata, Name: "token", Example: `"secret"`, Sensitive: true, Type: stringType()}}},
					AuthInfo:       &DataSchema{Members: []*Member{{Metadata: metadata, Name: "subject", Type: stringType()}}},
					PermEnabled:    true,
				},
			},
			{
				Metadata: metadata, Pub: false, Name: "Runtime", Kind: DeclarationTypeConfig, SkelName: "demo.contract.Runtime",
				Data: &DataSchema{Lifecycle: ConfigLifecycleEternal, Members: []*Member{{Metadata: metadata, Name: "endpoint", Type: stringType()}}},
			},
			{
				Metadata: metadata, Pub: true, Name: "Page", Kind: DeclarationTypeData, SkelName: "demo.contract.Page",
				Data: &DataSchema{
					Sensitive: true, TypeParameters: []string{"T"},
					Members: []*Member{
						{Metadata: metadata, Name: "item", Type: &Type{Kind: TypeKindTypeParameter, Name: "T"}},
						{Metadata: metadata, Name: "status", Type: &Type{Kind: TypeKindEnum, Name: "demo.contract.Status"}},
						{Metadata: metadata, Name: "parent", Type: &Type{Kind: TypeKindData, Name: "demo.contract.Page", Nullable: true, Arguments: []*Type{stringType()}}},
						{Metadata: metadata, Name: "runtime", Type: &Type{Kind: TypeKindConfig, Name: "demo.contract.Runtime"}},
						{Metadata: metadata, Name: "changed", Type: &Type{Kind: TypeKindEvent, Name: "demo.contract.UserChanged"}},
						{Metadata: metadata, Name: "external", Type: &Type{Kind: TypeKindImportedReference, Name: "external.users.User", Arguments: []*Type{intType()}}},
						{Metadata: metadata, Name: "index", Type: &Type{Kind: TypeKindMap, Key: stringType(), Value: &Type{Kind: TypeKindList, Element: &Type{Kind: TypeKindImportedReference, Name: "external.users.User", Nullable: true}}}},
					},
				},
			},
			{
				Metadata: metadata, Pub: true, Name: "Status", Kind: DeclarationTypeEnum, SkelName: "demo.contract.Status",
				Enum: &EnumSchema{Items: []*EnumItem{{Metadata: metadata, Name: "ACTIVE"}, {Name: "DISABLED"}}},
			},
			{
				Metadata: metadata, Pub: true, Name: "UserChanged", Kind: DeclarationTypeEvent, SkelName: "demo.contract.UserChanged",
				Data: &DataSchema{Members: []*Member{{Metadata: metadata, Name: "id", Type: stringType()}}},
			},
			{
				Metadata: metadata, Pub: true, Name: "User", Kind: DeclarationTypeResource, SkelName: "demo.contract.User",
				Resource: &ResourceSchema{
					Checks: []*ResourceCheck{{Metadata: metadata, Name: "owns", Arguments: []*Argument{argument("userId", stringType())}}},
					Actions: []*ResourceAction{{
						Metadata: metadata, Name: "read", PermissionCode: "demo.contract.User:read",
						Checks: []*ResourceCheck{{Metadata: metadata, Name: "allowed", Arguments: []*Argument{argument("permission", scalarType("string"))}}},
					}},
				},
			},
			{
				Metadata: metadata, Pub: true, Name: "Users", Kind: DeclarationTypeService, SkelName: "demo.contract.Users",
				Service: &ServiceSchema{
					Audiences: []*Audience{{Actor: "demo.contract.Caller", Via: "http"}}, Auth: AuthModeAuth,
					Require: &Requirement{Mode: RequirementModeAll, Children: []*Requirement{
						{Mode: RequirementModeCode, Code: "demo.contract.User:read"},
						{Mode: RequirementModeAny, Children: []*Requirement{
							{Mode: RequirementModeReference, Check: &RequirementCheck{Resource: "demo.contract.User", Action: "read", Check: "allowed", Arguments: []*RequirementCheckArgument{{JSONPath: "$.userId", Type: stringType()}}}},
							{Mode: RequirementModeCheck, Check: &RequirementCheck{Resource: "demo.contract.User", Check: "owns", Arguments: []*RequirementCheckArgument{{Name: "userId", JSONPath: "$.userId", Type: stringType()}}}},
						}},
					}},
					Methods: []*Method{{
						Metadata: metadata, Name: "get", SkelName: "get", Example: "get example", Auth: AuthModeNoAuth,
						Require:          &Requirement{Mode: RequirementModeCode, Code: "demo.contract.User:read"},
						InputDescription: "Input docs.", ArgumentsSensitive: true,
						OutputDescription: "Output docs.", OutputExample: "output example", ResultSensitive: true,
						Arguments: []*Argument{argument("id", stringType())},
						Result:    &Type{Kind: TypeKindData, Name: "demo.contract.Page", Arguments: []*Type{&Type{Kind: TypeKindImportedReference, Name: "external.users.User"}}},
					}},
				},
			},
			{
				Metadata: metadata, Pub: false, Name: "Jobs", Kind: DeclarationTypeTask, SkelName: "demo.contract.Jobs",
				Task: &TaskSchema{Triggers: []*Trigger{{
					Metadata: metadata, Name: "sync", SkelName: "sync", InputDescription: "Trigger input.", ArgumentsSensitive: true,
					Arguments: []*Argument{argument("cursor", stringType())},
				}}},
			},
			{
				Metadata: metadata, Pub: false, Name: "Dashboard", Kind: DeclarationTypeWeb, SkelName: "demo.contract.Dashboard",
				Web: &WebSchema{Audiences: []*Audience{{Actor: "demo.contract.Caller", Via: "http"}}},
			},
		},
	}
}
