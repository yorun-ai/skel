package schema

import (
	"strings"
	"testing"
)

type _RuleCoverage struct {
	covered map[string]ImpactLevel
}

func (c *_RuleCoverage) assert(t *testing.T, changes []*Change, want map[string]ImpactLevel) {
	t.Helper()
	if len(changes) != len(want) {
		t.Fatalf("expected %d changes, got %d: %+v", len(want), len(changes), changes)
	}
	for _, change := range changes {
		impact, ok := want[change.Code]
		if !ok {
			t.Fatalf("unexpected change code %q", change.Code)
		}
		if change.Impact != impact {
			t.Fatalf("change %s: expected impact %s, got %s", change.Code, impact, change.Impact)
		}
		changeType := ChangeModified
		switch {
		case strings.HasSuffix(change.Code, ".added"):
			changeType = ChangeAdded
		case strings.HasSuffix(change.Code, ".removed"):
			changeType = ChangeRemoved
		}
		if change.Change != changeType {
			t.Fatalf("change %s: expected type %s, got %s", change.Code, changeType, change.Change)
		}
	}
	for code, impact := range want {
		if previous, exists := c.covered[code]; exists {
			t.Fatalf("stable change code %s covered twice (%s and %s)", code, previous, impact)
		}
		c.covered[code] = impact
	}
}

func diffChanges(compare func(*_Diff)) []*Change {
	diff := &_Diff{report: &Report{Changes: []*Change{}}}
	compare(diff)
	return diff.report.Changes
}

func newTestDocument(declarations ...*Declaration) *Document {
	return &Document{
		Format: Format, FormatVersion: FormatVersion, Domain: "demo.user",
		Declarations: append([]*Declaration{}, declarations...),
	}
}

func scalarType(name string) *Type {
	return &Type{Kind: TypeKindScalar, Name: name}
}

func dataDeclaration(name string, members ...string) *Declaration {
	values := make([]*Member, 0, len(members))
	for _, member := range members {
		values = append(values, &Member{Name: member, Type: scalarType("string")})
	}
	return &Declaration{
		Pub: true, Name: name, Kind: DeclarationTypeData, SkelName: "demo.user." + name,
		Data: &DataSchema{Members: values},
	}
}

func enumDeclaration(name string, items ...string) *Declaration {
	values := make([]*EnumItem, 0, len(items))
	for _, item := range items {
		values = append(values, &EnumItem{Name: item})
	}
	return &Declaration{
		Pub: true, Name: name, Kind: DeclarationTypeEnum, SkelName: "demo.user." + name,
		Enum: &EnumSchema{Items: values},
	}
}

func serviceDeclaration(name string, methods ...string) *Declaration {
	values := make([]*Method, 0, len(methods))
	for _, method := range methods {
		values = append(values, &Method{Name: method, SkelName: method, Auth: AuthModeUnset, Arguments: []*Argument{}})
	}
	return &Declaration{
		Pub: true, Name: name, Kind: DeclarationTypeService, SkelName: "demo.user." + name,
		Service: &ServiceSchema{Auth: AuthModeUnset, Audiences: []*Audience{}, Methods: values},
	}
}

func actorDocument(authEnabled, permEnabled bool) *Document {
	actor := &ActorSchema{Vias: []*ActorVia{}, AuthEnabled: authEnabled, PermEnabled: permEnabled}
	if authEnabled {
		actor.AuthCredential = &DataSchema{Members: []*Member{}}
		actor.AuthInfo = &DataSchema{Members: []*Member{}}
	}
	return newTestDocument(&Declaration{
		Pub: true, Name: "UserActor", Kind: DeclarationTypeActor, SkelName: "demo.user.UserActor", Actor: actor,
	})
}

func resourceDocument(permissionCode string) *Document {
	return newTestDocument(&Declaration{
		Pub: true, Name: "User", Kind: DeclarationTypeResource, SkelName: "demo.user.User",
		Resource: &ResourceSchema{Actions: []*ResourceAction{{
			Name: "read", PermissionCode: permissionCode, Checks: []*ResourceCheck{},
		}}},
	})
}

func servicePolicyDocument(auth AuthMode, require *Requirement) *Document {
	declaration := serviceDeclaration("UserService")
	declaration.Service.Auth = auth
	declaration.Service.Require = require
	return newTestDocument(declaration)
}
