package analyzer

import (
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
	"go.yorun.ai/skelc/internal/util/nameutil"
)

func parseRequire(reporter *_DiagnosticReporter, gr *grammar.Require) (*model.PermissionRequire, bool) {
	if gr == nil {
		return nil, true
	}
	expr, valid := parseRequireExpr(reporter, gr.Expr)
	return &model.PermissionRequire{
		Expr: expr,
	}, valid
}

func parseRequireExpr(reporter *_DiagnosticReporter, expr *grammar.RequireExpr) (*model.PermissionExpr, bool) {
	if reporter.cancelled() {
		return nil, false
	}
	if expr == nil {
		return nil, false
	}
	if expr.Term != nil {
		return &model.PermissionExpr{Check: parseRequireTerm(expr.Term)}, true
	}

	mode := model.PermissionRequireMode(expr.Mode)
	valid := reporter.check(mode == model.PermissionRequireModeAll || mode == model.PermissionRequireModeAny,
		"unsupported require mode %s", expr.Mode)
	children := make([]*model.PermissionExpr, 0, len(expr.Children))
	for _, child := range expr.Children {
		if reporter.cancelled() {
			break
		}
		parsed, childValid := parseRequireExpr(reporter, child)
		valid = childValid && valid
		if parsed != nil {
			children = append(children, parsed)
		}
	}
	return &model.PermissionExpr{
		Mode:     mode,
		Children: children,
	}, valid
}

func parseRequireTerm(term *grammar.RequireTerm) *model.PermissionCheckInvocation {
	resourceRef := term.Target.Resource.String()
	action := term.Target.Action.Value
	item := &model.PermissionCheckInvocation{
		ResourceSkelName: resourceRef,
		ActionName:       action,
	}
	if term.Call == nil {
		return item
	}
	item.CheckName = term.Call.Name.Value
	item.Arguments = make([]*model.PermissionCheckArgument, 0, len(term.Call.Arguments))
	for _, arg := range term.Call.Arguments {
		item.Arguments = append(item.Arguments, &model.PermissionCheckArgument{
			JsonPath: arg.String(),
		})
	}
	return item
}

func (p *Analysis) normalizeServiceTypes(service *model.Service, refs *_RefContext) bool {
	valid := true
	for _, method := range service.Methods {
		if p.reporter.cancelled() {
			break
		}
		valid = fixTypeRef(p.reporter, method.ResultType, refs) && valid
		for _, arg := range method.Arguments {
			if p.reporter.cancelled() {
				break
			}
			valid = fixTypeRef(p.reporter, arg.Type, refs) && valid
		}
	}
	return valid
}

func (p *Analysis) normalizeServiceRequire(service *model.Service) bool {
	valid := p.normalizeRequire(service.Require, false, nil, service.Pos)
	for _, method := range service.Methods {
		if p.reporter.cancelled() {
			break
		}
		valid = p.normalizeRequire(method.Require, true, method, method.Pos) && valid
	}
	return valid
}

func (p *Analysis) normalizeRequire(require *model.PermissionRequire, allowChecks bool, method *model.Method, ownerPos model.Position) bool {
	if require == nil {
		return true
	}
	expr, valid := p.normalizeRequireExpr(require.Expr, allowChecks, method, ownerPos)
	if valid {
		require.Expr = expr
	}
	return valid
}

func (p *Analysis) normalizeRequireExpr(expr *model.PermissionExpr, allowChecks bool, method *model.Method, ownerPos model.Position) (*model.PermissionExpr, bool) {
	if p.reporter.cancelled() {
		return nil, false
	}
	if expr.Mode == "" && expr.Check != nil {
		return p.normalizeRequireItem(expr.Check, allowChecks, method, ownerPos)
	}

	valid := p.reporter.check(len(expr.Children) > 0, "%s require %s must have at least one item", ownerPos, expr.Mode)
	children := make([]*model.PermissionExpr, 0, len(expr.Children))
	for _, child := range expr.Children {
		if p.reporter.cancelled() {
			break
		}
		normalized, childValid := p.normalizeRequireExpr(child, allowChecks, method, ownerPos)
		valid = childValid && valid
		if normalized != nil {
			children = append(children, normalized)
		}
	}
	expr.Children = children
	return expr, valid
}

func (p *Analysis) normalizeRequireItem(item *model.PermissionCheckInvocation, allowChecks bool, method *model.Method, ownerPos model.Position) (*model.PermissionExpr, bool) {
	resourceRef := item.ResourceSkelName
	resource, action := p.findResourceAction(resourceRef, item.ActionName)
	if !p.reporter.checkReference(resource != nil, `%s require references undefined resource "%s"`, ownerPos, resourceRef) {
		return nil, false
	}
	if !p.reporter.checkReference(action != nil, `%s require references undefined action "%s"`, ownerPos, item.ActionName) {
		return nil, false
	}
	if !p.reporter.check(!p.isImportedResourceRef(resourceRef) || resource.Pub,
		`%s require references non-pub resource "%s"`, ownerPos, resourceRef) {
		return nil, false
	}
	codeExpr := &model.PermissionExpr{
		Mode: model.PermissionRequireModeCode,
		Code: action.PermissionCode,
	}
	if item.CheckName == "" {
		return codeExpr, true
	}

	if !p.reporter.check(allowChecks, "%s service require does not support check method", ownerPos) {
		return nil, false
	}
	check := findResourceCheck(resource, action, item.CheckName)
	if !p.reporter.checkReference(check != nil, `%s require references undefined check "%s"`, ownerPos, item.CheckName) {
		return nil, false
	}
	checkArguments := resourceCheckArguments(check)
	if !p.reporter.check(len(item.Arguments) == len(checkArguments),
		"%s require check %s expects %d argument(s), got %d",
		ownerPos, item.CheckName, len(checkArguments), len(item.Arguments)) {
		return nil, false
	}
	valid := true
	for index, argument := range item.Arguments {
		if p.reporter.cancelled() {
			break
		}
		checkArgument := checkArguments[index]
		argument.Name = checkArgument.Name
		argument.Type = checkArgument.Type
		valueType, pathValid := resolveMethodArgumentJsonPath(p.reporter, method, argument.JsonPath)
		if !pathValid {
			valid = false
			continue
		}
		valid = p.reporter.check(typeEqual(valueType, checkArgument.Type),
			"%s require check argument %s expects %s, got %s from %s",
			ownerPos, checkArgument.Name, checkArgument.Type.Name(), valueType.Name(), argument.JsonPath) && valid
	}
	return &model.PermissionExpr{
		Mode: model.PermissionRequireModeAll,
		Children: []*model.PermissionExpr{
			codeExpr,
			{
				Mode: model.PermissionRequireModeCheck,
				Check: &model.PermissionCheckInvocation{
					ResourceSkelName: resource.SkelName,
					ActionName:       item.ActionName,
					CheckName:        item.CheckName,
					ServiceSkelName:  resource.CheckService.SkelName,
					MethodSkelName:   check.Method.SkelName,
					CodeArgumentName: check.Method.Arguments[0].Name,
					Arguments:        item.Arguments,
				},
			},
		},
	}, valid
}

func resourceCheckArguments(check *model.ResourceCheck) []*model.Argument {
	if len(check.Method.Arguments) > 0 && check.Method.Arguments[0].Source == model.ArgumentSourcePermissionCode {
		return check.Method.Arguments[1:]
	}
	return check.Method.Arguments
}

func (p *Analysis) isImportedResourceRef(resourceRef string) bool {
	qualifier, _, ok := nameutil.SplitQualified(resourceRef)
	return ok && p.importsMap[qualifier] != nil
}

func (p *Analysis) findResourceAction(resourceRef string, actionName string) (*model.Resource, *model.ResourceAction) {
	resource := p.resourceByRef(resourceRef)
	if resource == nil {
		return nil, nil
	}
	for _, action := range resource.Actions {
		if p.reporter.cancelled() {
			break
		}
		if action.Name == actionName {
			return resource, action
		}
	}
	return resource, nil
}

func findResourceCheck(resource *model.Resource, action *model.ResourceAction, checkName string) *model.ResourceCheck {
	for _, check := range resource.Checks {
		if check.Name == checkName {
			return check
		}
	}
	for _, check := range action.Checks {
		if check.Name == checkName {
			return check
		}
	}
	return nil
}
