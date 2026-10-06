package codegen

import (
	"fmt"

	"go.yorun.ai/skel/internal/util/webpath"
	"go.yorun.ai/skel/schema"
)

// ValidateDomain rejects malformed programmatically constructed models before
// target renderers dereference tagged union fields or convert enum values.
func ValidateDomain(domain *schema.Domain) error {
	return validateDomain(domain, map[*schema.Domain]bool{})
}

func validateDomain(domain *schema.Domain, seen map[*schema.Domain]bool) error {
	if domain == nil {
		return fmt.Errorf("cannot generate code for a nil domain")
	}
	if seen[domain] {
		return nil
	}
	seen[domain] = true
	for _, domainImport := range domain.Imports() {
		if domainImport == nil {
			return fmt.Errorf("generated schema contains nil import")
		}
		if domainImport.Domain == nil {
			return fmt.Errorf("import %s has no domain schema", domainImport.Name)
		}
		if err := validateDomain(domainImport.Domain, seen); err != nil {
			return fmt.Errorf("import %s: %w", domainImport.Name, err)
		}
	}
	for _, values := range [][]*schema.Data{domain.Data(), domain.Configs(), domain.Events()} {
		for _, data := range values {
			if err := validateData(data); err != nil {
				return err
			}
		}
	}
	for _, enum := range domain.Enums() {
		if enum == nil {
			return fmt.Errorf("generated schema contains nil enum")
		}
		if enum.UnspecifiedItem == nil {
			return fmt.Errorf("enum %s has no unspecified item", enum.Name)
		}
		for _, item := range enum.Items {
			if item == nil {
				return fmt.Errorf("enum %s contains a nil item", enum.Name)
			}
		}
	}
	for _, config := range domain.Configs() {
		if config.Lifecycle != schema.ConfigLifecycleEternal && config.Lifecycle != schema.ConfigLifecycleInstant {
			return fmt.Errorf("config %s has unsupported lifecycle %q", config.Name, config.Lifecycle)
		}
	}
	for _, actor := range domain.Actors() {
		if err := validateActor(actor); err != nil {
			return err
		}
	}
	for _, service := range domain.Services() {
		if err := validateService(service); err != nil {
			return err
		}
	}
	for _, resource := range domain.Resources() {
		if err := validateResource(resource); err != nil {
			return err
		}
	}
	for _, web := range domain.Webs() {
		if web == nil {
			return fmt.Errorf("generated schema contains nil web")
		}
		if web.Auth != schema.AuthModeOff {
			if err := validateAuthMode(web.Auth); err != nil {
				return fmt.Errorf("web %s: %w", web.Name, err)
			}
		}
		if web.MountPath != "" {
			if err := webpath.Validate(web.MountPath); err != nil {
				return fmt.Errorf("web %s has invalid mount path: %w", web.Name, err)
			}
		}
		if err := validateAudiences("web "+web.Name, web.Audiences); err != nil {
			return err
		}
	}
	for _, task := range domain.Tasks() {
		if task == nil {
			return fmt.Errorf("generated schema contains nil task")
		}
		for _, trigger := range task.Triggers {
			if trigger == nil {
				return fmt.Errorf("task %s contains a nil trigger", task.Name)
			}
			if err := validateArguments("task trigger "+trigger.Name, trigger.Arguments, trigger.ArgumentsData); err != nil {
				return err
			}
		}
	}
	return nil
}
