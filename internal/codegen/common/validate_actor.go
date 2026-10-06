package common

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

func validateActor(actor *model.Actor) error {
	if actor == nil {
		return fmt.Errorf("generated model contains nil actor")
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
					valid = kind.Kind == model.TypeKindScalar && !kind.Nullable && (kind.Scalar == model.ScalarString || kind.Scalar == model.ScalarUUID || kind.Scalar == model.ScalarInt)
				}
			}
		}
		if !valid {
			return fmt.Errorf("actor %s identifier must name a non-nullable string, uuid, or int info field", actor.Name)
		}
	}
	if actor.PermEnabled {
		if actor.PermService == nil || actor.PermMethod == nil {
			return fmt.Errorf("actor %s has incomplete permission support", actor.Name)
		}
	}
	if actor.PermService != nil {
		if actor.PermService.Api {
			return fmt.Errorf("API service %s cannot be used as a framework callback", actor.PermService.Name)
		}
		if err := validateService(actor.PermService); err != nil {
			return fmt.Errorf("actor %s permission: %w", actor.Name, err)
		}
	}
	if actor.PermMethod != nil {
		if err := validateMethod("actor "+actor.Name+" permission method", actor.PermMethod); err != nil {
			return err
		}
	}
	return nil
}

func validateAudiences(owner string, audiences []*model.ActorAudience) error {
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
	switch model.ActorViaKind(via) {
	case model.ActorViaClient, model.ActorViaAgent, model.ActorViaOpenAPI:
		return nil
	default:
		return fmt.Errorf("unsupported actor via %q", via)
	}
}
