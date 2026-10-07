package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen/binding/golang/view"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

const sensitiveMarkerMethodName = "SkelSensitive"

// validateSensitiveMembers checks only structures emitted by this Go view.
// Field names that collide with generated methods are not language-level errors.
func (g *_Gen) validateSensitiveMembers() error {
	for _, data := range g.view.Data {
		if err := validateSensitiveData(data); err != nil {
			return err
		}
	}
	if g.mode == view.ModeApi {
		return nil
	}
	for _, data := range g.view.Configs {
		if err := validateSensitiveData(data); err != nil {
			return err
		}
	}
	for _, event := range g.view.Events {
		if g.isSplitRegular() && event.Ext {
			continue // The payload is an alias of the public package's type.
		}
		if err := validateSensitiveData(event); err != nil {
			return err
		}
	}
	for _, actor := range g.authServiceActors() {
		if actor.Auth == nil {
			continue
		}
		for _, data := range []*schema.Data{actor.Auth.Credential, actor.Auth.Info} {
			if err := validateSensitiveData(data); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSensitiveData(data *schema.Data) error {
	if !data.Sensitive {
		return nil
	}
	for _, member := range data.Members {
		if nameutil.ToCamel(member.Name) != sensitiveMarkerMethodName {
			continue
		}
		name := data.SkelName
		if name == "" {
			name = data.Name
		}
		message := fmt.Sprintf("Go field %s.%s conflicts with generated sensitive marker method %s", name, member.Name, sensitiveMarkerMethodName)
		if member.Pos.Line > 0 {
			return fmt.Errorf("%s %s", member.Pos, message)
		}
		return fmt.Errorf("%s", message)
	}
	return nil
}
