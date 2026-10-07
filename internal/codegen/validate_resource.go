package codegen

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

func validateResource(resource *schema.Resource) error {
	if resource == nil {
		return fmt.Errorf("generated schema contains nil resource")
	}
	if err := validateResourceChecks("resource "+resource.Name, resource.CheckService, resource.Checks); err != nil {
		return err
	}
	for _, action := range resource.Actions {
		if action == nil {
			return fmt.Errorf("resource %s contains a nil action", resource.Name)
		}
		if err := validateResourceChecks("resource "+resource.Name+" action "+action.Name, resource.CheckService, action.Checks); err != nil {
			return err
		}
	}
	if resource.CheckService != nil {
		if resource.CheckService.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", resource.CheckService.Name)
		}
		if err := validateService(resource.CheckService); err != nil {
			return err
		}
	}
	return nil
}

func validateResourceChecks(owner string, service *schema.Service, checks []*schema.ResourceCheck) error {
	for _, check := range checks {
		if check == nil {
			return fmt.Errorf("%s contains a nil check", owner)
		}
		if err := validateServiceMethodReference(owner+" check "+check.Name, service, check.Method); err != nil {
			return err
		}
	}
	return nil
}
