package common

import (
	"fmt"

	"go.yorun.ai/skelc/internal/model"
)

func validateService(service *model.Service) error {
	if service == nil {
		return fmt.Errorf("generated model contains nil service")
	}
	if service.Ext && (service.Pub || service.Api) {
		return fmt.Errorf("ext, api and pub are mutually exclusive")
	}
	if service.Api && service.Pub {
		return fmt.Errorf("service %s cannot combine api and pub", service.Name)
	}
	if service.Api && len(service.Audiences) == 0 {
		return fmt.Errorf("API service %s must declare at least one for Actor", service.Name)
	}
	if err := validateAuthMode(service.Auth); err != nil {
		return fmt.Errorf("service %s: %w", service.Name, err)
	}
	if err := validatePermissionExpr(service.Require); err != nil {
		return fmt.Errorf("service %s: %w", service.Name, err)
	}
	if err := validateAudiences("service "+service.Name, service.Audiences); err != nil {
		return err
	}
	for _, method := range service.Methods {
		if method == nil {
			return fmt.Errorf("service %s contains a nil method", service.Name)
		}
		if err := validateMethod("service "+service.Name+" method "+method.Name, method); err != nil {
			return err
		}
	}
	return nil
}

func validateMethod(owner string, method *model.Method) error {
	if method == nil {
		return fmt.Errorf("%s is nil", owner)
	}
	if err := validateAuthMode(method.Auth); err != nil {
		return fmt.Errorf("%s: %w", owner, err)
	}
	if err := validatePermissionExpr(method.Require); err != nil {
		return fmt.Errorf("%s: %w", owner, err)
	}
	if err := validateArguments(owner, method.Arguments, method.ArgumentsData); err != nil {
		return err
	}
	if method.ResultType != nil {
		if err := validateModelType(method.ResultType); err != nil {
			return fmt.Errorf("%s result: %w", owner, err)
		}
	}
	return nil
}

func validateArguments(owner string, arguments []*model.Argument, data *model.Data) error {
	members := map[string]bool{}
	if data != nil {
		if err := validateData(data); err != nil {
			return fmt.Errorf("%s arguments: %w", owner, err)
		}
		for _, member := range data.Members {
			members[member.Name] = true
		}
	}
	for _, argument := range arguments {
		if argument == nil {
			return fmt.Errorf("%s contains a nil argument", owner)
		}
		if err := validateModelType(argument.Type); err != nil {
			return fmt.Errorf("%s argument %s: %w", owner, argument.Name, err)
		}
		if data != nil && !members[argument.Name] {
			return fmt.Errorf("%s argument member %s not found", owner, argument.Name)
		}
	}
	return nil
}

func validateAuthMode(mode model.AuthMode) error {
	switch mode {
	case "", model.AuthModeUnset, model.AuthModeAuth, model.AuthModeNoAuth:
		return nil
	default:
		return fmt.Errorf("unsupported auth mode %q", mode)
	}
}
