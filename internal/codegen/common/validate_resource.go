package common

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

func validateResource(resource *model.Resource) error {
	if resource == nil {
		return fmt.Errorf("generated model contains nil resource")
	}
	if err := validateResourceChecks("resource "+resource.Name, resource.Checks); err != nil {
		return err
	}
	for _, action := range resource.Actions {
		if action == nil {
			return fmt.Errorf("resource %s contains a nil action", resource.Name)
		}
		if err := validateResourceChecks("resource "+resource.Name+" action "+action.Name, action.Checks); err != nil {
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

func validateResourceChecks(owner string, checks []*model.ResourceCheck) error {
	for _, check := range checks {
		if check == nil {
			return fmt.Errorf("%s contains a nil check", owner)
		}
		if err := validateMethod(owner+" check "+check.Name, check.Method); err != nil {
			return err
		}
	}
	return nil
}
