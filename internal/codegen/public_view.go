package codegen

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/projection"
	"go.yorun.ai/skel/internal/schema"
)

// PublicView contains declarations that belong to a domain's public contract.
type PublicView struct {
	Enums     []*model.Enum
	Data      []*model.Data
	Configs   []*model.Data
	Actors    []*model.Actor
	Resources []*model.Resource
	Events    []*model.Data
	Services  []*model.Service
}

// ProjectSchema projects selected declarations without changing their original
// public attributes. Imports remain available for canonical reference names.
func (v *PublicView) ProjectSchema(domain *model.Domain) (*schema.Document, error) {
	selected := model.NewDomainFromSpec(model.DomainSpec{
		Name: domain.Name(), Description: domain.Description(), Imports: domain.Imports(),
		Enums: v.Enums, Data: v.Data, Configs: v.Configs, Actors: v.Actors,
		Resources: v.Resources, Events: v.Events, Services: v.Services,
	})
	return projection.Project(selected, nil)
}

// BuildPublicView constructs and validates one public-contract projection.
func BuildPublicView(domain *model.Domain) (*PublicView, error) {
	view := &PublicView{
		Enums:     filter(domain.Enums(), func(value *model.Enum) bool { return value.Pub }),
		Data:      filter(domain.Data(), func(value *model.Data) bool { return value.Pub }),
		Configs:   filter(domain.Configs(), func(value *model.Data) bool { return value.Pub }),
		Actors:    filter(domain.Actors(), func(value *model.Actor) bool { return value.Pub }),
		Resources: filter(domain.Resources(), func(value *model.Resource) bool { return value.Pub }),
		Events:    filter(domain.Events(), func(value *model.Data) bool { return value.Public() }),
		Services:  filter(domain.Services(), func(value *model.Service) bool { return value.Public() }),
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

func validatePublicView(domain *model.Domain, view *PublicView) error {
	publicResources := make(map[string]bool, len(view.Resources))
	for _, resource := range view.Resources {
		publicResources[resource.SkelName] = true
	}
	for _, data := range view.Data {
		if err := validateMembers("pub data "+data.Name, data.Members, map[*model.Data]bool{}); err != nil {
			return err
		}
	}
	for _, config := range view.Configs {
		if err := validateMembers("pub config "+config.Name, config.Members, map[*model.Data]bool{}); err != nil {
			return err
		}
	}
	for _, actor := range view.Actors {
		if !actor.AuthEnabled {
			continue
		}
		if actor.AuthCredential == nil || actor.AuthInfo == nil {
			return fmt.Errorf("pub actor %s has incomplete auth data", actor.Name)
		}
		if err := validateMembers("pub actor "+actor.Name+" credential", actor.AuthCredential.Members, map[*model.Data]bool{}); err != nil {
			return err
		}
		if err := validateMembers("pub actor "+actor.Name+" info", actor.AuthInfo.Members, map[*model.Data]bool{}); err != nil {
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
				if err := validateType(context, argument.Type, map[*model.Data]bool{}); err != nil {
					return err
				}
			}
			if err := validateType(context, method.ResultType, map[*model.Data]bool{}); err != nil {
				return err
			}
		}
	}
	for _, event := range view.Events {
		if err := validateMembers("pub event "+event.Name, event.Members, map[*model.Data]bool{}); err != nil {
			return err
		}
	}
	return nil
}

func validateMembers(context string, members []*model.DataMember, visited map[*model.Data]bool) error {
	for _, member := range members {
		if err := validateType(context, member.Type, visited); err != nil {
			return err
		}
	}
	return nil
}

func validateType(context string, valueType *model.Type, visited map[*model.Data]bool) error {
	if valueType == nil {
		return nil
	}
	switch valueType.Kind {
	case model.TypeKindEnum:
		if valueType.ExternalDomain != "" && valueType.Enum != nil && !valueType.Enum.Pub {
			return fmt.Errorf("%s references non-pub enum %s", context, valueType.Enum.Name)
		}
	case model.TypeKindData:
		if valueType.Data == nil {
			return nil
		}
		if valueType.ExternalDomain != "" && valueType.Data.Kind == model.DataKindData && !valueType.Data.Pub {
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
	case model.TypeKindList:
		if valueType.List == nil {
			return fmt.Errorf("%s contains an invalid list type", context)
		}
		return validateType(context, valueType.List.Value, visited)
	case model.TypeKindMap:
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

func validateRequire(domainName, context string, require *model.PermissionRequire, publicResources map[string]bool) error {
	if require == nil {
		return nil
	}
	return validateRequireExpr(domainName, context, require.Expr, publicResources)
}

func validateRequireExpr(domainName, context string, expr *model.PermissionExpr, publicResources map[string]bool) error {
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

func findActor(actors []*model.Actor, name string) *model.Actor {
	for _, actor := range actors {
		if actor.Name == name {
			return actor
		}
	}
	return nil
}
