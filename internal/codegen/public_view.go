package codegen

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/schema"
)

// PublicView contains declarations that belong to a domain's public contract.
type PublicView struct {
	Enums     []*schema.Enum
	Data      []*schema.Data
	Configs   []*schema.Data
	Actors    []*schema.Actor
	Resources []*schema.Resource
	Events    []*schema.Data
	Services  []*schema.Service
}

// Schema returns a domain view borrowing the selected semantic declarations.
// It retains source metadata and imports; it does not modify the original graph.
func (v *PublicView) Schema(domain *schema.Domain) *schema.Domain {
	return schema.NewDomainFromSpec(schema.DomainSpec{
		Name: domain.Name(), Description: domain.Description(), Hash: domain.Hash(), Imports: domain.Imports(),
		Enums: v.Enums, Data: v.Data, Configs: v.Configs, Actors: v.Actors,
		Resources: v.Resources, Events: v.Events, Services: v.Services,
	})
}

// BuildPublicView constructs and validates one public-contract projection.
func BuildPublicView(domain *schema.Domain) (*PublicView, error) {
	view := &PublicView{
		Enums:     filter(domain.Enums(), func(value *schema.Enum) bool { return value.Pub }),
		Data:      filter(domain.Data(), func(value *schema.Data) bool { return value.Pub }),
		Configs:   filter(domain.Configs(), func(value *schema.Data) bool { return value.Pub }),
		Actors:    filter(domain.Actors(), func(value *schema.Actor) bool { return value.Pub }),
		Resources: filter(domain.Resources(), func(value *schema.Resource) bool { return value.Pub }),
		Events:    filter(domain.Events(), func(value *schema.Data) bool { return value.Pub || value.Ext }),
		Services:  filter(domain.Services(), func(value *schema.Service) bool { return value.Pub || value.Ext }),
	}
	collectViewData(domain, view)
	if err := validatePublicView(domain, view); err != nil {
		return nil, err
	}
	return view, nil
}

func filter[T any](values []*T, keep func(*T) bool) []*T {
	filtered := make([]*T, 0, len(values))
	for _, value := range values {
		if keep(value) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func validatePublicView(domain *schema.Domain, view *PublicView) error {
	publicResources := make(map[string]bool, len(view.Resources))
	for _, resource := range view.Resources {
		publicResources[resource.SkelName] = true
	}
	for _, data := range view.Data {
		if err := validateMembers("pub data "+data.Name, data.Members, map[*schema.Data]bool{}); err != nil {
			return err
		}
	}
	for _, config := range view.Configs {
		if err := validateMembers("pub config "+config.Name, config.Members, map[*schema.Data]bool{}); err != nil {
			return err
		}
	}
	for _, actor := range view.Actors {
		if actor.Auth == nil {
			continue
		}
		if actor.Auth.Credential == nil || actor.Auth.Info == nil {
			return fmt.Errorf("pub actor %s has incomplete auth data", actor.Name)
		}
		if err := validateMembers("pub actor "+actor.Name+" credential", actor.Auth.Credential.Members, map[*schema.Data]bool{}); err != nil {
			return err
		}
		if err := validateMembers("pub actor "+actor.Name+" info", actor.Auth.Info.Members, map[*schema.Data]bool{}); err != nil {
			return err
		}
	}
	for _, service := range view.Services {
		for _, audience := range service.Audiences {
			if actor := findActor(domain.Actors(), audience.Actor); actor != nil && !actor.Pub {
				return fmt.Errorf("pub service %s references non-pub actor %s", service.Name, actor.Name)
			}
		}
		if err := validateRequire(domain.Name(), "pub service "+service.Name, service.Require, publicResources); err != nil {
			return err
		}
		for _, method := range service.Methods {
			context := fmt.Sprintf("pub service %s.%s", service.Name, method.Name)
			if err := validateRequire(domain.Name(), context, method.Require, publicResources); err != nil {
				return err
			}
			for _, argument := range method.Arguments {
				if err := validateType(context, argument.Type, map[*schema.Data]bool{}); err != nil {
					return err
				}
			}
			if err := validateType(context, method.ResultType, map[*schema.Data]bool{}); err != nil {
				return err
			}
		}
	}
	for _, event := range view.Events {
		if err := validateMembers("pub event "+event.Name, event.Members, map[*schema.Data]bool{}); err != nil {
			return err
		}
	}
	return nil
}

func validateMembers(context string, members []*schema.DataMember, visited map[*schema.Data]bool) error {
	for _, member := range members {
		if err := validateType(context, member.Type, visited); err != nil {
			return err
		}
	}
	return nil
}

func validateType(context string, valueType *schema.Type, visited map[*schema.Data]bool) error {
	if valueType == nil {
		return nil
	}
	switch valueType.Kind {
	case schema.TypeKindEnum:
		if valueType.ExternalDomain != "" && valueType.Enum != nil && !valueType.Enum.Pub {
			return fmt.Errorf("%s references non-pub enum %s", context, valueType.Enum.Name)
		}
	case schema.TypeKindData:
		if valueType.Data == nil {
			return nil
		}
		if valueType.ExternalDomain != "" && valueType.Data.Kind == schema.DataKindData && !valueType.Data.Pub {
			return fmt.Errorf("%s references non-pub data %s", context, valueType.Data.Name)
		}
		if visited[valueType.Data] {
			return nil
		}
		visited[valueType.Data] = true
		for _, argument := range valueType.TypeArguments {
			if err := validateType(context, argument, visited); err != nil {
				return err
			}
		}
		return validateMembers(context, valueType.Data.Members, visited)
	case schema.TypeKindList:
		if valueType.List == nil {
			return fmt.Errorf("%s contains an invalid list type", context)
		}
		return validateType(context, valueType.List.Element, visited)
	case schema.TypeKindMap:
		if valueType.Map == nil {
			return fmt.Errorf("%s contains an invalid map type", context)
		}
		if err := validateType(context, valueType.Map.Key, visited); err != nil {
			return err
		}
		return validateType(context, valueType.Map.Value, visited)
	}
	return nil
}

func validateRequire(domainName, context string, require *schema.PermissionRequire, publicResources map[string]bool) error {
	if require == nil {
		return nil
	}
	return validateRequireExpr(domainName, context, require.Expression, publicResources)
}

func validateRequireExpr(domainName, context string, expr *schema.PermissionExpression, publicResources map[string]bool) error {
	if expr == nil {
		return nil
	}
	if expr.Code != "" {
		index := strings.LastIndex(expr.Code, ":")
		if index <= 0 {
			return fmt.Errorf("invalid permission code %s", expr.Code)
		}
		resourceName := expr.Code[:index]
		if strings.HasPrefix(resourceName, domainName+".") && !publicResources[resourceName] {
			return fmt.Errorf("%s references non-pub resource %s", context, resourceName)
		}
	}
	for _, child := range expr.Children {
		if err := validateRequireExpr(domainName, context, child, publicResources); err != nil {
			return err
		}
	}
	return nil
}

func findActor(actors []*schema.Actor, name string) *schema.Actor {
	for _, actor := range actors {
		if actor.Name == name {
			return actor
		}
	}
	return nil
}
