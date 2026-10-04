package schema

import "fmt"

func validateService(value *ServiceSchema) error {
	if value.Audiences == nil {
		return fmt.Errorf("service audiences are required")
	}
	if value.Methods == nil {
		return fmt.Errorf("service methods are required")
	}
	if err := validateAuth(value.Auth); err != nil {
		return fmt.Errorf("service: %w", err)
	}
	if err := validateAudiences(value.Audiences); err != nil {
		return err
	}
	if err := validateRequirement(value.Require); err != nil {
		return err
	}
	seen := map[string]bool{}
	for index, method := range value.Methods {
		if method == nil || method.Name == "" || method.SkelName == "" {
			return fmt.Errorf("service method %d is null or unnamed", index)
		}
		if seen[method.Name] {
			return fmt.Errorf("duplicated service method %s", method.Name)
		}
		if method.Arguments == nil {
			return fmt.Errorf("service method %s arguments are required", method.Name)
		}
		if err := validateAuth(method.Auth); err != nil {
			return fmt.Errorf("service method %s: %w", method.Name, err)
		}
		seen[method.Name] = true
		if err := validateArguments(method.Arguments); err != nil {
			return fmt.Errorf("service method %s: %w", method.Name, err)
		}
		if err := validateTypeOptional(method.Result); err != nil {
			return fmt.Errorf("service method %s result: %w", method.Name, err)
		}
		if err := validateRequirement(method.Require); err != nil {
			return fmt.Errorf("service method %s: %w", method.Name, err)
		}
	}
	return nil
}

func validateAuth(value AuthMode) error {
	switch value {
	case AuthModeUnset, AuthModeAuth, AuthModeNoAuth, AuthModeRequired, AuthModeOptional, AuthModeGuest:
		return nil
	default:
		return fmt.Errorf("unsupported authentication mode %q", value)
	}
}

func validateArguments(values []*Argument) error {
	seen := map[string]bool{}
	for index, argument := range values {
		if argument == nil || argument.Name == "" {
			return fmt.Errorf("argument %d is null or unnamed", index)
		}
		if seen[argument.Name] {
			return fmt.Errorf("duplicated argument %s", argument.Name)
		}
		seen[argument.Name] = true
		if err := validateType(argument.Type); err != nil {
			return fmt.Errorf("argument %s: %w", argument.Name, err)
		}
	}
	return nil
}

func validateAudiences(values []*Audience) error {
	seen := map[string]bool{}
	for index, audience := range values {
		if audience == nil || audience.Actor == "" {
			return fmt.Errorf("audience %d is null or unnamed", index)
		}
		key := audience.Actor + "\x00" + audience.Via
		if seen[key] {
			return fmt.Errorf("duplicated audience %s via %s", audience.Actor, audience.Via)
		}
		seen[key] = true
	}
	return nil
}

func validateTask(value *TaskSchema) error {
	if value.Triggers == nil {
		return fmt.Errorf("task triggers are required")
	}
	seen := map[string]bool{}
	for index, trigger := range value.Triggers {
		if trigger == nil || trigger.Name == "" || trigger.SkelName == "" {
			return fmt.Errorf("task trigger %d is null or unnamed", index)
		}
		if seen[trigger.Name] {
			return fmt.Errorf("duplicated task trigger %s", trigger.Name)
		}
		if trigger.Arguments == nil {
			return fmt.Errorf("task trigger %s arguments are required", trigger.Name)
		}
		seen[trigger.Name] = true
		if err := validateArguments(trigger.Arguments); err != nil {
			return fmt.Errorf("task trigger %s: %w", trigger.Name, err)
		}
	}
	return nil
}
