package codegen

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

func validateActor(actor *schema.Actor) error {
	if actor == nil {
		return fmt.Errorf("generated schema contains nil actor")
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
		if len(actor.Auth.Info.Members) == 0 {
			return fmt.Errorf("actor %s auth info must have at least one member", actor.Name)
		}
		if actor.Auth.Result != nil {
			if err := validateAuthResult(actor); err != nil {
				return err
			}
		} else if len(actor.Auth.Impersonators) != 0 || actor.Auth.ImpersonatorInfo != nil {
			return fmt.Errorf("actor %s impersonators require an auth result", actor.Name)
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
	}
	return nil
}

func validateAuthResult(actor *schema.Actor) error {
	auth := actor.Auth
	if err := validateData(auth.Result); err != nil {
		return fmt.Errorf("actor %s auth result: %w", actor.Name, err)
	}
	resultType := auth.Method.ResultType
	if resultType == nil || resultType.Kind != schema.TypeKindData || resultType.Nullable || resultType.Data != auth.Result {
		return fmt.Errorf("actor %s auth method must return its auth result", actor.Name)
	}
	members := auth.Result.Members
	expected := 1
	if len(auth.Impersonators) > 0 {
		expected = 2
	}
	if len(members) != expected || members[0].Name != "info" || members[0].Type.Kind != schema.TypeKindData || members[0].Type.Nullable || members[0].Type.Data != auth.Info {
		return fmt.Errorf("actor %s auth result must contain its required info", actor.Name)
	}
	if expected == 1 {
		if auth.ImpersonatorInfo != nil {
			return fmt.Errorf("actor %s has undeclared impersonator info", actor.Name)
		}
		return nil
	}
	if auth.ImpersonatorInfo == nil || members[1].Name != "impersonatorInfo" || members[1].Type.Kind != schema.TypeKindData || !members[1].Type.Nullable || members[1].Type.Data != auth.ImpersonatorInfo || len(auth.ImpersonatorInfo.Members) != len(auth.Impersonators) {
		return fmt.Errorf("actor %s has invalid impersonator info", actor.Name)
	}
	fields, actors := map[string]bool{}, map[string]bool{}
	for index, ref := range auth.Impersonators {
		if ref == nil || ref.ActorSkelName == "" || ref.InfoFieldName == "" || fields[ref.InfoFieldName] || actors[ref.ActorSkelName] {
			return fmt.Errorf("actor %s has invalid or duplicate impersonator mapping", actor.Name)
		}
		fields[ref.InfoFieldName], actors[ref.ActorSkelName] = true, true
		member := auth.ImpersonatorInfo.Members[index]
		if member.Name != ref.InfoFieldName || member.Type.Kind != schema.TypeKindData || !member.Type.Nullable {
			return fmt.Errorf("actor %s impersonator %s must map to nullable info", actor.Name, ref.ActorSkelName)
		}
	}
	return nil
}
