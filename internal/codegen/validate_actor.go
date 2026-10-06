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
	if actor.AuthEnabled {
		if actor.AuthCredential == nil || actor.AuthInfo == nil || actor.AuthService == nil || actor.AuthMethod == nil {
			return fmt.Errorf("actor %s has incomplete auth support", actor.Name)
		}
		if err := validateData(actor.AuthCredential); err != nil {
			return fmt.Errorf("actor %s auth credential: %w", actor.Name, err)
		}
		if err := validateData(actor.AuthInfo); err != nil {
			return fmt.Errorf("actor %s auth info: %w", actor.Name, err)
		}
		if actor.AuthService.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", actor.AuthService.Name)
		}
		if err := validateService(actor.AuthService); err != nil {
			return fmt.Errorf("actor %s auth: %w", actor.Name, err)
		}
		if err := validateMethod("actor "+actor.Name+" auth method", actor.AuthMethod); err != nil {
			return err
		}
	}
	if actor.IdentifierField != "" {
		valid := false
		if actor.AuthEnabled && actor.AuthInfo != nil {
			for _, member := range actor.AuthInfo.Members {
				if member.Name == actor.IdentifierField {
					kind := member.Type
					valid = kind.Kind == schema.TypeKindScalar && !kind.Nullable && (kind.Scalar == schema.ScalarString || kind.Scalar == schema.ScalarUUID || kind.Scalar == schema.ScalarInt)
				}
			}
		}
		if !valid {
			return fmt.Errorf("actor %s identifier must name a non-nullable string, uuid, or int info field", actor.Name)
		}
	}
	if actor.PermissionEnabled {
		if actor.PermissionService == nil || actor.PermissionMethod == nil {
			return fmt.Errorf("actor %s has incomplete permission support", actor.Name)
		}
	}
	if actor.PermissionService != nil {
		if actor.PermissionService.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", actor.PermissionService.Name)
		}
		if err := validateService(actor.PermissionService); err != nil {
			return fmt.Errorf("actor %s permission: %w", actor.Name, err)
		}
	}
	if actor.PermissionMethod != nil {
		if err := validateMethod("actor "+actor.Name+" permission method", actor.PermissionMethod); err != nil {
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
