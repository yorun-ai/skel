package schema

import (
	"fmt"
	"reflect"

	"go.yorun.ai/skel/internal/policy"
)

// EffectivePolicy is the derived policy for a method in its enclosing service.
// Permission references retain the input graph's resolution state.
type EffectivePolicy struct {
	AuthMode AuthMode
	Require  *PermissionRequire
}

// ComputeEffectivePolicy derives authentication and permission requirements
// without modifying declarations or consulting their cached effective fields.
func ComputeEffectivePolicy(service *Service, method *Method) (EffectivePolicy, error) {
	if service == nil || method == nil {
		return EffectivePolicy{}, fmt.Errorf("effective policy requires a service and method")
	}
	auth, err := policy.EffectiveAuth(string(service.AuthMode), string(method.AuthMode))
	if err != nil {
		return EffectivePolicy{}, err
	}
	require, err := ComposeRequirements(service.Require, method.Require)
	if err != nil {
		return EffectivePolicy{}, err
	}
	return EffectivePolicy{AuthMode: AuthMode(auth), Require: require}, nil
}

// ComposeRequirements returns the conjunction of two requirements, expanding
// unresolved source terms while preserving check order and reference bindings.
// Permission nodes and argument values are copied; semantic type links are borrowed.
func ComposeRequirements(service, method *PermissionRequire) (*PermissionRequire, error) {
	roots := [2]*PermissionExpression{}
	for index, require := range []*PermissionRequire{service, method} {
		if require != nil {
			if require.Expression == nil {
				return nil, fmt.Errorf("permission requirement has no expression")
			}
			roots[index] = require.Expression
		}
	}
	expression, err := permissionExpressions.Conjoin(roots[0], roots[1])
	if err != nil || expression == nil {
		return nil, err
	}
	return new(PermissionRequire{Expression: expression}), nil
}

// ExpandPermissionExpression expands a source resource/action term into its
// code and optional check. It does not resolve imports or modify its input.
// Returned nodes may borrow the input's check binding and must be read-only.
func ExpandPermissionExpression(value *PermissionExpression) *PermissionExpression {
	if value == nil || value.Mode != "" || value.Check == nil {
		return value
	}
	check := value.Check
	code := new(PermissionExpression{Mode: PermissionRequireModeCode, Code: check.ResourceSkelName + ":" + check.ActionName})
	if check.CheckName == "" {
		return code
	}
	return new(PermissionExpression{Mode: PermissionRequireModeAll, Children: []*PermissionExpression{
		code, {Mode: PermissionRequireModeCheck, Check: check},
	}})
}

var permissionExpressions = policy.Expressions[PermissionExpression]{
	Expand:   ExpandPermissionExpression,
	Children: func(value *PermissionExpression) []*PermissionExpression { return value.Children },
	All: func(children []*PermissionExpression) *PermissionExpression {
		return new(PermissionExpression{Mode: PermissionRequireModeAll, Children: children})
	},
	Copy: func(value *PermissionExpression, children []*PermissionExpression) (*PermissionExpression, error) {
		switch value.Mode {
		case PermissionRequireModeCode, PermissionRequireModeCheck, PermissionRequireModeAll, PermissionRequireModeAny:
		default:
			return nil, fmt.Errorf("unsupported permission require mode %q", value.Mode)
		}
		copy := *value
		copy.Children = children
		if value.Check != nil {
			check := *value.Check
			if check.Arguments != nil {
				check.Arguments = make([]*PermissionCheckArgument, len(value.Check.Arguments))
				for index, argument := range value.Check.Arguments {
					if argument == nil {
						return nil, fmt.Errorf("permission check contains a nil argument")
					}
					check.Arguments[index] = new(PermissionCheckArgument(*argument))
				}
			}
			copy.Check = &check
		} else if value.Mode == PermissionRequireModeCheck {
			return nil, fmt.Errorf("permission check invocation is nil")
		}
		return &copy, nil
	},
}

// PopulateEffectivePolicies refreshes the derived fields of methods owned by
// domain, including actor and resource callbacks. It does not visit imports.
// Call after editing declarations and before sharing the graph read-only.
// On error, no method is changed.
func PopulateEffectivePolicies(domain *Domain) error {
	values := map[*Method]EffectivePolicy{}
	err := visitPolicyMethods(domain, func(service *Service, method *Method) error {
		value, err := ComputeEffectivePolicy(service, method)
		if err == nil {
			if previous, exists := values[method]; exists && !reflect.DeepEqual(previous, value) {
				return fmt.Errorf("method is shared by services with different effective policies")
			}
			values[method] = value
		}
		return err
	})
	if err != nil {
		return err
	}
	for method, value := range values {
		method.EffectiveAuthMode, method.EffectiveRequire = value.AuthMode, value.Require
	}
	return nil
}

// ValidateEffectivePolicy checks only whether stored derived fields match the
// declarations. It never populates or repairs values, resolves imports, or
// performs general schema validation. All owned method policies are checked.
func ValidateEffectivePolicy(domain *Domain) error {
	return visitPolicyMethods(domain, func(service *Service, method *Method) error {
		expected, err := ComputeEffectivePolicy(service, method)
		if err != nil {
			return err
		}
		if method.EffectiveAuthMode != expected.AuthMode {
			return fmt.Errorf("effectiveAuthMode: got %q, expected %q", method.EffectiveAuthMode, expected.AuthMode)
		}
		if !reflect.DeepEqual(method.EffectiveRequire, expected.Require) {
			return fmt.Errorf("effectiveRequire does not match declared requirements")
		}
		return nil
	})
}

func visitPolicyMethods(domain *Domain, visit func(*Service, *Method) error) error {
	if domain == nil {
		return fmt.Errorf("effective policy requires a domain")
	}
	services := append([]*Service(nil), domain.Services()...)
	for _, actor := range domain.Actors() {
		if actor == nil {
			continue
		}
		if actor.Auth != nil {
			services = append(services, actor.Auth.Service)
		}
		if actor.Permission != nil {
			services = append(services, actor.Permission.Service)
		}
	}
	for _, resource := range domain.Resources() {
		if resource != nil && resource.CheckService != nil {
			services = append(services, resource.CheckService)
		}
	}
	for _, service := range services {
		if service == nil {
			return fmt.Errorf("cannot derive policy for nil service")
		}
		for _, method := range service.Methods {
			if method == nil {
				return fmt.Errorf("service %s contains nil method", service.Name)
			}
			path := "services." + service.Name + ".methods." + method.Name
			if err := visit(service, method); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
	}
	return nil
}
