package projection

import (
	"fmt"
	"slices"
	"strings"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/schema"
)

func Project(domain *model.Domain, importAliases map[string]string) (*schema.Document, error) {
	if domain == nil {
		return nil, fmt.Errorf("cannot project a nil domain")
	}
	document := &schema.Document{
		Format: schema.Format, FormatVersion: schema.FormatVersion, Domain: domain.Name(),
		Description: domain.Description(), Declarations: []*schema.Declaration{},
	}
	aliases := referenceAliases(domain, importAliases)
	appendDeclarations(document, domain, aliases, domain.Enums(), domain.Data(), domain.Configs(), domain.Events(),
		domain.Actors(), domain.Resources(), domain.Services(), domain.Webs(), domain.Tasks())
	normalizeReferenceNames(document, domain.Name(), aliases)
	slices.SortFunc(document.Declarations, compareDeclarations)
	return document, nil
}

// ProjectDataDeclaration normalizes one semantic data-like declaration for an
// internal consumer that needs the same representation as a schema document.
func ProjectDataDeclaration(domain *model.Domain, value *model.Data) *schema.Declaration {
	if value == nil {
		return nil
	}
	projected := projectData(value)
	if domain != nil {
		normalizeDeclarationReferences(projected, domain.Name(), referenceAliases(domain, nil))
	}
	return projected
}

// ProjectServiceDeclaration normalizes one semantic service declaration for an
// internal consumer that needs the same representation as a schema document.
func ProjectServiceDeclaration(domain *model.Domain, value *model.Service) *schema.Declaration {
	if value == nil {
		return nil
	}
	domainName := ""
	aliases := map[string]string{}
	if domain != nil {
		domainName = domain.Name()
		aliases = referenceAliases(domain, nil)
	}
	declaration := projectService(domainName, aliases, value)
	normalizeDeclarationReferences(declaration, domainName, aliases)
	return declaration
}

// ProjectMethodSchema normalizes one semantic method for an internal consumer
// that needs the same representation as a schema document.
func ProjectMethodSchema(domain *model.Domain, value *model.Method) *schema.Method {
	if value == nil {
		return nil
	}
	projected := projectMethod(value)
	if domain != nil {
		domainName := domain.Name()
		aliases := referenceAliases(domain, nil)
		normalizeArgumentReferences(projected.Arguments, domainName, aliases)
		normalizeTypeReference(projected.Result, domainName, aliases)
		normalizeRequirementReferences(projected.Require, domainName, aliases)
	}
	return projected
}

func referenceAliases(domain *model.Domain, supplied map[string]string) map[string]string {
	aliases := make(map[string]string, len(supplied)+len(domain.Imports()))
	for alias, name := range supplied {
		aliases[alias] = name
	}
	for _, imported := range domain.Imports() {
		aliases[imported.Alias] = imported.Name
	}
	return aliases
}

func appendDeclarations(
	document *schema.Document,
	domain *model.Domain,
	importAliases map[string]string,
	enums []*model.Enum,
	data []*model.Data,
	configs []*model.Data,
	events []*model.Data,
	actors []*model.Actor,
	resources []*model.Resource,
	services []*model.Service,
	webs []*model.Web,
	tasks []*model.Task,
) {
	for _, value := range enums {
		document.Declarations = append(document.Declarations, projectEnum(value))
	}
	for _, value := range data {
		document.Declarations = append(document.Declarations, projectData(value))
	}
	for _, value := range configs {
		document.Declarations = append(document.Declarations, projectData(value))
	}
	for _, value := range events {
		document.Declarations = append(document.Declarations, projectData(value))
	}
	for _, value := range actors {
		document.Declarations = append(document.Declarations, projectActor(value))
	}
	for _, value := range resources {
		document.Declarations = append(document.Declarations, projectResource(value))
	}
	for _, value := range services {
		document.Declarations = append(document.Declarations, projectService(domain.Name(), importAliases, value))
	}
	for _, value := range webs {
		document.Declarations = append(document.Declarations, projectWeb(domain.Name(), importAliases, value))
	}
	for _, value := range tasks {
		document.Declarations = append(document.Declarations, projectTask(value))
	}
}

func projectEnum(value *model.Enum) *schema.Declaration {
	items := make([]*schema.EnumItem, 0, len(value.Items))
	for _, item := range value.Items {
		items = append(items, &schema.EnumItem{Metadata: metadata(item.Description, item.Deprecated, item.DeprecatedReason), Name: item.Name, Pos: item.Pos})
	}
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:      value.Pub, Name: value.Name, Kind: schema.DeclarationTypeEnum, SkelName: value.SkelName, Pos: value.Pos,
		Enum: &schema.EnumSchema{Items: items},
	}
}

func projectData(value *model.Data) *schema.Declaration {
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:      value.Public(), Name: value.Name, Kind: schema.DeclarationType(value.Kind), SkelName: value.SkelName, Pos: value.Pos,
		Data: projectDataSchema(value),
	}
}

func projectDataSchema(value *model.Data) *schema.DataSchema {
	if value == nil {
		return nil
	}
	typeParameters := make([]string, 0, len(value.TypeParameters))
	for _, parameter := range value.TypeParameters {
		typeParameters = append(typeParameters, parameter.Name)
	}
	return &schema.DataSchema{
		Lifecycle: schema.ConfigLifecycle(value.Lifecycle), Sensitive: value.Sensitive, Ext: value.Ext,
		TypeParameters: typeParameters, Members: projectMembers(value.Members),
	}
}

func projectMembers(values []*model.DataMember) []*schema.Member {
	members := make([]*schema.Member, 0, len(values))
	for _, value := range values {
		members = append(members, &schema.Member{
			Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
			Name:     value.Name, Example: value.Example, Sensitive: value.Sensitive, Type: projectType(value.Type), Pos: value.Pos,
		})
	}
	return members
}

func projectActor(value *model.Actor) *schema.Declaration {
	vias := make([]*schema.ActorVia, 0, len(value.Vias))
	for _, via := range value.Vias {
		vias = append(vias, &schema.ActorVia{Name: via.Name, Pos: via.Pos})
	}
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:      value.Pub, Name: value.Name, Kind: schema.DeclarationTypeActor, SkelName: value.SkelName, Pos: value.Pos,
		Actor: &schema.ActorSchema{
			IdentifierField: value.IdentifierField, Vias: vias, AuthEnabled: value.AuthEnabled, AuthCredential: projectDataSchema(value.AuthCredential),
			AuthInfo: projectDataSchema(value.AuthInfo), PermEnabled: value.PermEnabled,
		},
	}
}

func projectResource(value *model.Resource) *schema.Declaration {
	actions := make([]*schema.ResourceAction, 0, len(value.Actions))
	for _, action := range value.Actions {
		actions = append(actions, &schema.ResourceAction{
			Metadata: metadata(action.Description, action.Deprecated, action.DeprecatedReason),
			Name:     action.Name, PermissionCode: action.PermissionCode, Checks: projectResourceChecks(action.Checks), Pos: action.Pos,
		})
	}
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:      value.Pub, Name: value.Name, Kind: schema.DeclarationTypeResource, SkelName: value.SkelName, Pos: value.Pos,
		Resource: &schema.ResourceSchema{Checks: projectResourceChecks(value.Checks), Actions: actions},
	}
}

func projectResourceChecks(values []*model.ResourceCheck) []*schema.ResourceCheck {
	checks := make([]*schema.ResourceCheck, 0, len(values))
	for _, value := range values {
		checks = append(checks, &schema.ResourceCheck{
			Metadata: metadata(value.Method.Description, value.Deprecated, value.DeprecatedReason),
			Name:     value.Name, Arguments: projectArguments(value.Method.Arguments), Pos: value.Method.Pos,
		})
	}
	return checks
}

func projectService(domainName string, importAliases map[string]string, value *model.Service) *schema.Declaration {
	methods := make([]*schema.Method, 0, len(value.Methods))
	for _, method := range value.Methods {
		methods = append(methods, projectMethod(method))
	}
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:      value.Public(), Name: value.Name, Kind: schema.DeclarationTypeService, SkelName: value.SkelName, Pos: value.Pos,
		Service: &schema.ServiceSchema{
			Audiences: projectAudiences(domainName, importAliases, value.Audiences),
			Api:       value.Api,
			Ext:       value.Ext,
			Auth:      normalizedServiceAuth(value),
			Require:   projectRequirement(value.Require), Methods: methods,
		},
	}
}

func projectMethod(value *model.Method) *schema.Method {
	return &schema.Method{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:     value.Name, SkelName: value.SkelName, Example: value.Example, Auth: normalizedAuth(value.Auth),
		Require: projectRequirement(value.Require), InputDescription: value.InputDescription,
		ArgumentsSensitive: value.ArgumentsSensitive, OutputDescription: value.OutputDescription,
		OutputExample: value.OutputExample, ResultSensitive: value.ResultSensitive,
		Arguments: projectArguments(value.Arguments), Result: projectType(value.ResultType), Pos: value.Pos,
	}
}

func projectArguments(values []*model.Argument) []*schema.Argument {
	arguments := make([]*schema.Argument, 0, len(values))
	for _, value := range values {
		arguments = append(arguments, &schema.Argument{
			Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
			Name:     value.Name, Example: value.Example, Sensitive: value.Sensitive, Type: projectType(value.Type), Pos: value.Pos,
		})
	}
	return arguments
}

func projectWeb(domainName string, importAliases map[string]string, value *model.Web) *schema.Declaration {
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:     value.Name, Kind: schema.DeclarationTypeWeb, SkelName: value.SkelName, Pos: value.Pos,
		Web: &schema.WebSchema{Auth: normalizedWebAuth(value.Auth), Audiences: projectAudiences(domainName, importAliases, value.Audiences), MountPath: value.MountPath},
	}
}

func projectTask(value *model.Task) *schema.Declaration {
	triggers := make([]*schema.Trigger, 0, len(value.Triggers))
	for _, trigger := range value.Triggers {
		triggers = append(triggers, &schema.Trigger{
			Metadata: metadata(trigger.Description, trigger.Deprecated, trigger.DeprecatedReason),
			Name:     trigger.Name, SkelName: trigger.SkelName, InputDescription: trigger.InputDescription,
			ArgumentsSensitive: trigger.ArgumentsSensitive, Arguments: projectArguments(trigger.Arguments), Pos: trigger.Pos,
		})
	}
	return &schema.Declaration{
		Metadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:     value.Name, Kind: schema.DeclarationTypeTask, SkelName: value.SkelName, Pos: value.Pos,
		Task: &schema.TaskSchema{Triggers: triggers},
	}
}

func projectAudiences(domainName string, importAliases map[string]string, values []*model.ActorAudience) []*schema.Audience {
	audiences := make([]*schema.Audience, 0, len(values))
	for _, value := range values {
		audiences = append(audiences, &schema.Audience{Actor: canonicalReferenceName(domainName, importAliases, value.Actor), Via: value.Via, Pos: value.Pos})
	}
	return audiences
}

func compareDeclarations(left, right *schema.Declaration) int {
	leftOrder := kindOrder(left.Kind)
	rightOrder := kindOrder(right.Kind)
	if leftOrder != rightOrder {
		return leftOrder - rightOrder
	}
	return strings.Compare(left.SkelName, right.SkelName)
}

func kindOrder(kind schema.DeclarationType) int {
	switch kind {
	case schema.DeclarationTypeActor:
		return 1
	case schema.DeclarationTypeConfig:
		return 2
	case schema.DeclarationTypeData:
		return 3
	case schema.DeclarationTypeEnum:
		return 4
	case schema.DeclarationTypeEvent:
		return 5
	case schema.DeclarationTypeResource:
		return 6
	case schema.DeclarationTypeService:
		return 7
	case schema.DeclarationTypeTask:
		return 8
	case schema.DeclarationTypeWeb:
		return 9
	default:
		return 99
	}
}

func normalizedWebAuth(mode model.AuthMode) schema.AuthMode {
	switch mode {
	case "", model.AuthModeUnset, model.AuthModeAuth:
		return schema.AuthModeRequired
	case model.AuthModeNoAuth:
		return schema.AuthModeOff
	default:
		return schema.AuthMode(mode)
	}
}

func normalizedServiceAuth(service *model.Service) schema.AuthMode {
	if service.Auth == "" || service.Auth == model.AuthModeUnset {
		return schema.AuthModeRequired
	}
	return normalizedAuth(service.Auth)
}
