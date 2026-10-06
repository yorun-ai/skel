package descriptor

import (
	"fmt"
	"reflect"

	"go.yorun.ai/skel/internal/policy"
)

// EffectivePolicy is the derived policy for a method in its enclosing service.
type EffectivePolicy struct {
	AuthMode AuthMode
	Require  *PermissionRequire
}

// ComputeEffectivePolicy derives policy from declared fields without changing
// descriptors, evaluating requests, or consulting stored effective fields.
func ComputeEffectivePolicy(service *Service, method *Method) (EffectivePolicy, error) {
	if service == nil || method == nil {
		return EffectivePolicy{}, fmt.Errorf("effective policy requires a service and method")
	}
	auth, err := policy.EffectiveAuth(string(service.AuthMode), string(method.AuthMode))
	if err != nil {
		return EffectivePolicy{}, err
	}
	roots := [2]*PermissionExpression{}
	for index, require := range []*PermissionRequire{service.Require, method.Require} {
		if require != nil {
			if require.Expression == nil {
				return EffectivePolicy{}, fmt.Errorf("permission requirement has no expression")
			}
			roots[index] = require.Expression
		}
	}
	expression, err := permissionExpressions.Conjoin(roots[0], roots[1])
	if err != nil {
		return EffectivePolicy{}, err
	}
	var require *PermissionRequire
	if expression != nil {
		require = new(PermissionRequire{Expression: expression})
	}
	return EffectivePolicy{AuthMode: AuthMode(auth), Require: require}, nil
}

// ValidateEffectivePolicy checks only that each owned method's effective fields
// equal the policy derived from its declarations. This includes actor and
// resource callbacks. It never repairs fields, validates external targets,
// registers descriptors or evaluates requests. Consumers may call it at registration.
func ValidateEffectivePolicy(domain *Domain) error {
	if domain == nil {
		return fmt.Errorf("effective policy requires a domain")
	}
	services := append([]*Service(nil), domain.Services...)
	for _, actor := range domain.Actors {
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
	for _, resource := range domain.Resources {
		if resource != nil && resource.CheckService != nil {
			services = append(services, resource.CheckService)
		}
	}
	for _, service := range services {
		if service == nil {
			return fmt.Errorf("cannot verify effective policy of nil service")
		}
		for _, method := range service.Methods {
			if method == nil {
				return fmt.Errorf("service %s contains nil method", service.Name)
			}
			path := "services." + service.Name + ".methods." + method.Name
			expected, err := ComputeEffectivePolicy(service, method)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if method.EffectiveAuthMode != expected.AuthMode {
				return fmt.Errorf("%s.effectiveAuthMode: got %q, expected %q", path, method.EffectiveAuthMode, expected.AuthMode)
			}
			if !reflect.DeepEqual(method.EffectiveRequire, expected.Require) {
				return fmt.Errorf("%s.effectiveRequire does not match declared requirements", path)
			}
		}
	}
	return nil
}

var permissionExpressions = policy.Expressions[PermissionExpression]{
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
					copy := *argument
					copy.Type = clonePolicyType(copy.Type, map[*Type]*Type{})
					check.Arguments[index] = &copy
				}
			}
			copy.Check = &check
		} else if value.Mode == PermissionRequireModeCheck {
			return nil, fmt.Errorf("permission check invocation is nil")
		}
		return &copy, nil
	},
}

func clonePolicyType(value *Type, copies map[*Type]*Type) *Type {
	if value == nil {
		return nil
	}
	if existing := copies[value]; existing != nil {
		return existing
	}
	copy := *value
	copies[value] = &copy
	copy.Element = clonePolicyType(value.Element, copies)
	copy.Key = clonePolicyType(value.Key, copies)
	copy.Value = clonePolicyType(value.Value, copies)
	if value.TypeArguments != nil {
		copy.TypeArguments = make([]*Type, len(value.TypeArguments))
		for index, argument := range value.TypeArguments {
			copy.TypeArguments[index] = clonePolicyType(argument, copies)
		}
	}
	return &copy
}
