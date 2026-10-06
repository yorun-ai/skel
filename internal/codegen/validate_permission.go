package codegen

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

func validatePermissionExpr(require *model.PermissionRequire) error {
	if require == nil || require.Expr == nil {
		return nil
	}
	active := map[*model.PermissionExpr]uint8{}
	var validate func(*model.PermissionExpr) error
	validate = func(expr *model.PermissionExpr) error {
		if expr == nil {
			return fmt.Errorf("permission expression is nil")
		}
		if active[expr] == 1 {
			return fmt.Errorf("cyclic permission expression")
		}
		if active[expr] == 2 {
			return nil
		}
		active[expr] = 1
		switch expr.Mode {
		case model.PermissionRequireModeCode:
		case model.PermissionRequireModeCheck:
			if expr.Check == nil {
				return fmt.Errorf("permission check invocation is nil")
			}
			for _, argument := range expr.Check.Arguments {
				if argument == nil {
					return fmt.Errorf("permission check contains a nil argument")
				}
				if err := validateModelType(argument.Type); err != nil {
					return fmt.Errorf("permission check argument %s: %w", argument.Name, err)
				}
			}
		case model.PermissionRequireModeAll, model.PermissionRequireModeAny:
			for _, child := range expr.Children {
				if err := validate(child); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported permission require mode %q", expr.Mode)
		}
		active[expr] = 2
		return nil
	}
	return validate(require.Expr)
}
