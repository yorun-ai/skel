package analyzer

import (
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
)

func parseRequire(reporter *_DiagnosticReporter, gr *grammar.Require) (*schema.PermissionRequire, bool) {
	if gr == nil {
		return nil, true
	}
	expr, valid := parseRequireExpr(reporter, gr.Expr)
	return &schema.PermissionRequire{
		Expression: expr,
	}, valid
}

func parseRequireExpr(reporter *_DiagnosticReporter, expr *grammar.RequireExpr) (*schema.PermissionExpression, bool) {
	if reporter.cancelled() {
		return nil, false
	}
	if expr == nil {
		return nil, false
	}
	if expr.Term != nil {
		return &schema.PermissionExpression{Check: parseRequireTerm(expr.Term)}, true
	}

	mode := schema.PermissionRequireMode(expr.Mode)
	valid := reporter.check(mode == schema.PermissionRequireModeAll || mode == schema.PermissionRequireModeAny,
		"unsupported require mode %s", expr.Mode)
	children := make([]*schema.PermissionExpression, 0, len(expr.Children))
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
	return &schema.PermissionExpression{
		Mode:     mode,
		Children: children,
	}, valid
}

func parseRequireTerm(term *grammar.RequireTerm) *schema.PermissionCheckInvocation {
	resourceRef := term.Target.Resource.String()
	action := term.Target.Action.Value
	item := &schema.PermissionCheckInvocation{
		ResourceSkelName: resourceRef,
		ActionName:       action,
	}
	if term.Call == nil {
		return item
	}
	item.CheckName = term.Call.Name.Value
	item.Arguments = make([]*schema.PermissionCheckArgument, 0, len(term.Call.Arguments))
	for _, arg := range term.Call.Arguments {
		item.Arguments = append(item.Arguments, &schema.PermissionCheckArgument{
			JsonPath: arg.String(),
		})
	}
	return item
}

func (p *Analysis) normalizeServiceTypes(service *schema.Service, refs *_RefContext) bool {
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

func (p *Analysis) normalizeServiceRequire(service *schema.Service) bool {
	valid := p.normalizeRequire(service.Require, false, nil, service.Pos)
	for _, method := range service.Methods {
		if p.reporter.cancelled() {
			break
		}
		valid = p.normalizeRequire(method.Require, true, method, method.Pos) && valid
	}
	return valid
}

func (p *Analysis) normalizeRequire(require *schema.PermissionRequire, allowChecks bool, method *schema.Method, ownerPos schema.Position) bool {
	if require == nil {
		return true
	}
	expr, valid := p.normalizeRequireExpr(require.Expression, allowChecks, method, ownerPos)
	if valid {
		require.Expression = expr
	}
	return valid
}

func (p *Analysis) normalizeRequireExpr(expr *schema.PermissionExpression, allowChecks bool, method *schema.Method, ownerPos schema.Position) (*schema.PermissionExpression, bool) {
	if p.reporter.cancelled() {
		return nil, false
	}
	if expr.Mode == "" && expr.Check != nil {
		return p.normalizeRequireItem(expr.Check, allowChecks, method, ownerPos)
	}

	valid := p.reporter.check(len(expr.Children) > 0, "%s require %s must have at least one item", ownerPos, expr.Mode)
	children := make([]*schema.PermissionExpression, 0, len(expr.Children))
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

func (p *Analysis) normalizeRequireItem(item *schema.PermissionCheckInvocation, allowChecks bool, method *schema.Method, ownerPos schema.Position) (*schema.PermissionExpression, bool) {
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
	codeExpr := &schema.PermissionExpression{
		Mode: schema.PermissionRequireModeCode,
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
	return &schema.PermissionExpression{
		Mode: schema.PermissionRequireModeAll,
		Children: []*schema.PermissionExpression{
			codeExpr,
			{
				Mode: schema.PermissionRequireModeCheck,
				Check: &schema.PermissionCheckInvocation{
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

func resourceCheckArguments(check *schema.ResourceCheck) []*schema.Argument {
	if len(check.Method.Arguments) > 0 && check.Method.Arguments[0].Source == schema.ArgumentSourcePermissionCode {
		return check.Method.Arguments[1:]
	}
	return check.Method.Arguments
}

func (p *Analysis) isImportedResourceRef(resourceRef string) bool {
	qualifier, _, ok := nameutil.SplitQualified(resourceRef)
	return ok && p.importsMap[qualifier] != nil
}

func (p *Analysis) findResourceAction(resourceRef string, actionName string) (*schema.Resource, *schema.ResourceAction) {
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

func findResourceCheck(resource *schema.Resource, action *schema.ResourceAction, checkName string) *schema.ResourceCheck {
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
