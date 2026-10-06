package output

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yorun.ai/skel/api"
	"go.yorun.ai/skel/schema"
)

func TestSchemaOutputIncludesEffectivePolicies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.skel")
	result, err := api.QuerySchema(api.Input{SkelIn: path, Sources: map[string][]byte{path: []byte(`domain demo
import shared as dep
actor ClientActor { via client {} permission {} }
api service DocumentApiService {
    for ClientActor
    auth optional
    require dep.Document:read
    method get { require dep.Document:read:byId(id) input { id: string } }
}
web PortalWeb { for ClientActor via client auth off }
`)}}, api.SchemaQueryOption{})
	if err != nil {
		t.Fatal(err)
	}
	wire := DescribeSchemaDeclaration(result.Domain, result.Domain.Find(schema.DeclarationTypeService, "demo.DocumentApiService"))
	method := wire.Service.Methods[0]
	if method.AuthMode != "inherit" || method.EffectiveAuthMode != "optional" || method.Require == nil ||
		method.EffectiveRequire.Children[0].Code != "shared.Document:read" {
		t.Fatalf("CLI lost declared/effective policies: %+v", method)
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"authMode":"optional"`, `"authMode":"inherit"`, `"effectiveAuthMode":"optional"`, `"require":`, `"effectiveRequire":`} {
		if !strings.Contains(string(encoded), field) {
			t.Fatalf("CLI JSON missing %s: %s", field, encoded)
		}
	}
	web := DescribeSchemaDeclaration(result.Domain, result.Domain.Find(schema.DeclarationTypeWeb, "demo.PortalWeb"))
	encodedWeb, err := json.Marshal(web)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encodedWeb), `"authMode":"off"`) {
		t.Fatalf("web JSON lost authentication mode: %s", encodedWeb)
	}
	for _, content := range [][]byte{encoded, encodedWeb} {
		for _, old := range []string{`"auth":`, `"effectiveAuth":`} {
			if strings.Contains(string(content), old) {
				t.Fatalf("CLI JSON retains old mode field %s: %s", old, content)
			}
		}
	}
}

func TestSchemaOutputMapsSchemaEnums(t *testing.T) {
	t.Run("config lifecycle", func(t *testing.T) {
		tests := []struct {
			semantic schema.ConfigLifecycle
			wire     string
		}{
			{semantic: schema.ConfigLifecycleEternal, wire: "eternal"},
			{semantic: schema.ConfigLifecycleInstant, wire: "instant"},
		}
		for _, test := range tests {
			got := (&_SchemaEncoder{}).projectDataSchema(new(schema.Data{Lifecycle: test.semantic, Members: []*schema.DataMember{}})).Lifecycle
			if got != test.wire {
				t.Fatalf("lifecycle %q projects to %q, want %q", test.semantic, got, test.wire)
			}
		}
	})

	t.Run("authentication", func(t *testing.T) {
		tests := []struct {
			semantic schema.AuthMode
			wire     string
		}{
			{semantic: schema.AuthModeUnset, wire: "inherit"},
		}
		for _, test := range tests {
			if got := string((&schema.Method{AuthMode: test.semantic}).NormalizedAuth()); got != test.wire {
				t.Fatalf("auth %q projects to %q, want %q", test.semantic, got, test.wire)
			}
		}
	})
}

func TestSchemaOutputMapsTypeKinds(t *testing.T) {
	tests := []struct {
		name     string
		semantic *schema.Type
		wire     *SchemaType
	}{
		{name: "imported reference", semantic: new(schema.Type{Kind: schema.TypeKindUnresolvedReference, SkelName: "User", ExternalAlias: "shared"}), wire: new(SchemaType{Kind: "importedReference", Name: "shared.User"})},
		{name: "scalar", semantic: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarBoolean}), wire: new(SchemaType{Kind: "scalar", Name: "bool"})},
		{name: "list", semantic: new(schema.Type{Kind: schema.TypeKindList, List: new(schema.ListType{Element: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString})})}), wire: new(SchemaType{Kind: "list", Element: new(SchemaType{Kind: "scalar", Name: "string"})})},
		{name: "map", semantic: new(schema.Type{Kind: schema.TypeKindMap, Map: new(schema.MapType{Key: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString}), Value: new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarInt})})}), wire: new(SchemaType{Kind: "map", Key: new(SchemaType{Kind: "scalar", Name: "string"}), Value: new(SchemaType{Kind: "scalar", Name: "int"})})},
		{name: "enum", semantic: new(schema.Type{Kind: schema.TypeKindEnum, SkelName: "demo.contract.State"}), wire: new(SchemaType{Kind: "enum", Name: "demo.contract.State"})},
		{name: "data", semantic: new(schema.Type{Kind: schema.TypeKindData, SkelName: "demo.contract.Page", Nullable: true, TypeArguments: []*schema.Type{new(schema.Type{Kind: schema.TypeKindScalar, Scalar: schema.ScalarString})}}), wire: new(SchemaType{Kind: "data", Nullable: true, Name: "demo.contract.Page", Arguments: []*SchemaType{new(SchemaType{Kind: "scalar", Name: "string"})}})},
		{name: "config", semantic: new(schema.Type{Kind: schema.TypeKindData, Data: new(schema.Data{Kind: schema.DataKindConfig}), SkelName: "demo.contract.RuntimeConfig"}), wire: new(SchemaType{Kind: "config", Name: "demo.contract.RuntimeConfig"})},
		{name: "event", semantic: new(schema.Type{Kind: schema.TypeKindData, Data: new(schema.Data{Kind: schema.DataKindEvent}), SkelName: "demo.contract.ChangedEvent"}), wire: new(SchemaType{Kind: "event", Name: "demo.contract.ChangedEvent"})},
		{name: "type parameter", semantic: new(schema.Type{Kind: schema.TypeKindTypeParameter, TypeParameter: new(schema.TypeParameter{Name: "T"})}), wire: new(SchemaType{Kind: "typeParameter", Name: "T"})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := (&_SchemaEncoder{domain: schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo"})}).projectType(test.semantic); !reflect.DeepEqual(got, test.wire) {
				t.Fatalf("type projection mismatch:\nwant: %#v\ngot:  %#v", test.wire, got)
			}
		})
	}
}

func TestSchemaOutputMapsRequirementModes(t *testing.T) {
	check := new(schema.PermissionCheckInvocation{ResourceSkelName: "demo.contract.Document", CheckName: "owns"})
	tests := []struct {
		name     string
		semantic *schema.PermissionExpression
		wire     string
	}{
		{name: "code", semantic: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeCode, Code: "demo.contract.Document:read"}), wire: "code"},
		{name: "check", semantic: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeCheck, Check: check}), wire: "check"},
		{name: "all", semantic: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAll, Children: []*schema.PermissionExpression{}}), wire: "all"},
		{name: "any", semantic: new(schema.PermissionExpression{Mode: schema.PermissionRequireModeAny, Children: []*schema.PermissionExpression{}}), wire: "any"},
		{name: "reference", semantic: new(schema.PermissionExpression{Check: check}), wire: "reference"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := (&_SchemaEncoder{domain: schema.NewDomainFromSpec(schema.DomainSpec{Name: "demo"})}).projectRequirementExpr(test.semantic); got.Mode != test.wire {
				t.Fatalf("requirement mode = %q, want %q", got.Mode, test.wire)
			}
		})
	}
}
