package diff

import (
	"strings"
	"testing"

	"go.yorun.ai/skel/schema"
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

func newTestDomain(declarations ...*schema.Declaration) *schema.Domain {
	spec := schema.DomainSpec{Name: "demo.user"}
	for _, d := range declarations {
		switch {
		case d.Enum != nil:
			d.Enum.Name, d.Enum.SkelName, d.Enum.Pos = d.Name, d.SkelName, d.Pos
			d.Enum.Description, d.Enum.Deprecated, d.Enum.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason
			d.Enum.Pub = d.Pub
			spec.Enums = append(spec.Enums, d.Enum)
		case d.Data != nil:
			d.Data.Name, d.Data.SkelName, d.Data.Pos = d.Name, d.SkelName, d.Pos
			d.Data.Description, d.Data.Deprecated, d.Data.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason
			d.Data.Pub = d.Pub
			d.Data.Kind = schema.DataKind(d.Kind)
			switch d.Kind {
			case schema.DeclarationTypeConfig:
				spec.Configs = append(spec.Configs, d.Data)
			case schema.DeclarationTypeEvent:
				spec.Events = append(spec.Events, d.Data)
			default:
				spec.Data = append(spec.Data, d.Data)
			}
		case d.Actor != nil:
			d.Actor.Name, d.Actor.SkelName, d.Actor.Pos = d.Name, d.SkelName, d.Pos
			d.Actor.Description, d.Actor.Deprecated, d.Actor.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason
			d.Actor.Pub = d.Pub
			spec.Actors = append(spec.Actors, d.Actor)
		case d.Resource != nil:
			d.Resource.Name, d.Resource.SkelName, d.Resource.Pos = d.Name, d.SkelName, d.Pos
			d.Resource.Description, d.Resource.Deprecated, d.Resource.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason
			d.Resource.Pub = d.Pub
			spec.Resources = append(spec.Resources, d.Resource)
		case d.Service != nil:
			d.Service.Name, d.Service.SkelName, d.Service.Pos = d.Name, d.SkelName, d.Pos
			d.Service.Description, d.Service.Deprecated, d.Service.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason
			d.Service.Pub = d.Pub
			spec.Services = append(spec.Services, d.Service)
		case d.Web != nil:
			d.Web.Name, d.Web.SkelName, d.Web.Pos = d.Name, d.SkelName, d.Pos
			d.Web.Description, d.Web.Deprecated, d.Web.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason

			spec.Webs = append(spec.Webs, d.Web)
		case d.Task != nil:
			d.Task.Name, d.Task.SkelName, d.Task.Pos = d.Name, d.SkelName, d.Pos
			d.Task.Description, d.Task.Deprecated, d.Task.DeprecatedReason = d.Description, d.Deprecated, d.DeprecatedReason

			spec.Tasks = append(spec.Tasks, d.Task)

		}
	}
	return schema.NewDomainFromSpec(spec)
}

func scalarType(name string) *schema.Type {
	values := map[string]schema.Scalar{"string": schema.ScalarString, "int": schema.ScalarInt, "bool": schema.ScalarBoolean}
	return &schema.Type{Kind: schema.TypeKindScalar, Scalar: values[name]}
}

func dataDeclaration(name string, members ...string) *schema.Declaration {
	values := make([]*schema.DataMember, 0, len(members))
	for _, member := range members {
		values = append(values, &schema.DataMember{Name: member, Type: scalarType("string")})
	}
	return &schema.Declaration{
		Pub: true, Name: name, Kind: schema.DeclarationTypeData, SkelName: "demo.user." + name,
		Data: &schema.Data{Members: values},
	}
}

func enumDeclaration(name string, items ...string) *schema.Declaration {
	values := make([]*schema.EnumItem, 0, len(items))
	for _, item := range items {
		values = append(values, &schema.EnumItem{Name: item})
	}
	return &schema.Declaration{
		Pub: true, Name: name, Kind: schema.DeclarationTypeEnum, SkelName: "demo.user." + name,
		Enum: &schema.Enum{Items: values},
	}
}

func serviceDeclaration(name string, methods ...string) *schema.Declaration {
	values := make([]*schema.Method, 0, len(methods))
	for _, method := range methods {
		values = append(values, &schema.Method{Name: method, SkelName: method, Auth: schema.AuthModeUnset, Arguments: []*schema.Argument{}})
	}
	return &schema.Declaration{
		Pub: true, Name: name, Kind: schema.DeclarationTypeService, SkelName: "demo.user." + name,
		Service: &schema.Service{Auth: schema.AuthModeUnset, Audiences: []*schema.ActorAudience{}, Methods: values},
	}
}

func actorDomain(authEnabled, permEnabled bool) *schema.Domain {
	actor := &schema.Actor{Vias: []*schema.ActorVia{}}
	if permEnabled {
		actor.Permission = new(schema.ActorPermission{})
	}
	if authEnabled {
		actor.Auth = new(schema.ActorAuth{})
		actor.Auth.Credential = &schema.Data{Members: []*schema.DataMember{}}
		actor.Auth.Info = &schema.Data{Members: []*schema.DataMember{}}
	}
	return newTestDomain(&schema.Declaration{
		Pub: true, Name: "UserActor", Kind: schema.DeclarationTypeActor, SkelName: "demo.user.UserActor", Actor: actor,
	})
}

func resourceDomain(permissionCode string) *schema.Domain {
	return newTestDomain(&schema.Declaration{
		Pub: true, Name: "User", Kind: schema.DeclarationTypeResource, SkelName: "demo.user.User",
		Resource: &schema.Resource{Actions: []*schema.ResourceAction{{
			Name: "read", PermissionCode: permissionCode, Checks: []*schema.ResourceCheck{},
		}}},
	})
}

func servicePolicyDomain(auth schema.AuthMode, require *schema.PermissionRequire) *schema.Domain {
	declaration := serviceDeclaration("UserService")
	declaration.Service.Auth = auth
	declaration.Service.Require = require
	return newTestDomain(declaration)
}

func requirementExpressions(values []*schema.PermissionRequire) []*schema.PermissionExpression {
	result := make([]*schema.PermissionExpression, 0, len(values))
	for _, value := range values {
		result = append(result, value.Expression)
	}
	return result
}
