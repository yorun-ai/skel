package codegen

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

func validateActor(actor *schema.Actor) error {
	if actor == nil {
		return fmt.Errorf("generated schema contains nil actor")
	}
	for _, via := range actor.Vias {
		if via == nil {
			return fmt.Errorf("actor %s contains a nil via", actor.Name)
		}
		if err := validateActorVia(via.Name); err != nil {
			return fmt.Errorf("actor %s: %w", actor.Name, err)
		}
	}
	if actor.Auth != nil {
		if actor.Auth.Credential == nil || actor.Auth.Info == nil || actor.Auth.Service == nil || actor.Auth.Method == nil {
			return fmt.Errorf("actor %s has incomplete auth support", actor.Name)
		}
		if err := validateData(actor.Auth.Credential); err != nil {
			return fmt.Errorf("actor %s auth credential: %w", actor.Name, err)
		}
		if err := validateData(actor.Auth.Info); err != nil {
			return fmt.Errorf("actor %s auth info: %w", actor.Name, err)
		}
		if actor.Auth.Service.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", actor.Auth.Service.Name)
		}
		if err := validateService(actor.Auth.Service); err != nil {
			return fmt.Errorf("actor %s auth: %w", actor.Name, err)
		}
		if err := validateServiceMethodReference("actor "+actor.Name+" auth method", actor.Auth.Service, actor.Auth.Method); err != nil {
			return err
		}
		if actor.Auth.IdentifierField != "" {
			valid := false
			for _, member := range actor.Auth.Info.Members {
				if member.Name == actor.Auth.IdentifierField {
					kind := member.Type
					valid = kind.Kind == schema.TypeKindScalar && !kind.Nullable && (kind.Scalar == schema.ScalarString || kind.Scalar == schema.ScalarUUID || kind.Scalar == schema.ScalarInt)
				}
			}
			if !valid {
				return fmt.Errorf("actor %s identifier must name a non-nullable string, uuid, or int info field", actor.Name)
			}
		}
	}
	if actor.Permission != nil {
		if actor.Permission.Service == nil || actor.Permission.Method == nil {
			return fmt.Errorf("actor %s has incomplete permission support", actor.Name)
		}
		if actor.Permission.Service.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", actor.Permission.Service.Name)
		}
		if err := validateService(actor.Permission.Service); err != nil {
			return fmt.Errorf("actor %s permission: %w", actor.Name, err)
		}
		if err := validateServiceMethodReference("actor "+actor.Name+" permission method", actor.Permission.Service, actor.Permission.Method); err != nil {
			return err
		}
	}
	return nil
}

func validateAudiences(owner string, audiences []*schema.ActorAudience) error {
	for _, audience := range audiences {
		if audience == nil {
			return fmt.Errorf("%s contains a nil audience", owner)
		}
		if audience.Actor == "" {
			return fmt.Errorf("%s contains an audience without an actor", owner)
		}
		if audience.Via != "" {
			if err := validateActorVia(audience.Via); err != nil {
				return fmt.Errorf("%s: %w", owner, err)
			}
		}
	}
	return nil
}

func validateActorVia(via string) error {
	switch schema.ActorViaKind(via) {
	case schema.ActorViaClient, schema.ActorViaAgent, schema.ActorViaOpenAPI:
		return nil
	default:
		return fmt.Errorf("unsupported actor via %q", via)
	}
}
