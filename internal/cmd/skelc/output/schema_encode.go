package output

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/schema"
)

// SchemaEntries serializes declaration summaries in stable schema order.
func SchemaEntries(domain *schema.Domain) []*SchemaEntry {
	result := make([]*SchemaEntry, 0)
	for _, value := range domain.Declarations() {
		result = append(result, new(SchemaEntry{Pub: value.Pub, Name: value.Name, Kind: value.Kind, SkelName: value.SkelName}))
	}
	return result
}

// DescribeSchemaDeclaration serializes one declaration without following graph links.
// Missing declarations are represented by nil and encode as JSON null.
func DescribeSchemaDeclaration(domain *schema.Domain, value *schema.Declaration) *SchemaDeclaration {
	if value == nil {
		return nil
	}
	encoder := _SchemaEncoder{domain: domain}
	var result *SchemaDeclaration
	switch {
	case value.Enum != nil:
		result = encoder.projectEnum(value.Enum)
	case value.Data != nil:
		result = encoder.projectData(value.Data)
	case value.Actor != nil:
		result = encoder.projectActor(value.Actor)
	case value.Resource != nil:
		result = encoder.projectResource(value.Resource)
	case value.Service != nil:
		result = encoder.projectService(value.Service)
	case value.Web != nil:
		result = encoder.projectWeb(value.Web)
	case value.Task != nil:
		result = encoder.projectTask(value.Task)
	}
	return result
}

func (e *_SchemaEncoder) projectEnum(value *schema.Enum) *SchemaDeclaration {
	items := make([]*SchemaEnumItem, 0, len(value.Items))
	for _, item := range value.Items {
		items = append(items, &SchemaEnumItem{SchemaMetadata: metadata(item.Description, item.Deprecated, item.DeprecatedReason), Name: item.Name})
	}
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:            value.Pub, Name: value.Name, Kind: schema.DeclarationTypeEnum, SkelName: value.SkelName,
		Enum: &SchemaEnum{Items: items},
	}
}

func (e *_SchemaEncoder) projectData(value *schema.Data) *SchemaDeclaration {
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:            value.Public(), Name: value.Name, Kind: schema.DeclarationType(value.Kind), SkelName: value.SkelName,
		Data: e.projectDataSchema(value),
	}
}

func (e *_SchemaEncoder) projectDataSchema(value *schema.Data) *SchemaData {
	if value == nil {
		return nil
	}
	typeParameters := make([]string, 0, len(value.TypeParameters))
	for _, parameter := range value.TypeParameters {
		typeParameters = append(typeParameters, parameter.Name)
	}
	return &SchemaData{
		Lifecycle: string(value.Lifecycle), Sensitive: value.Sensitive, Ext: value.Ext,
		TypeParameters: typeParameters, Members: e.projectMembers(value.Members),
	}
}

func (e *_SchemaEncoder) projectMembers(values []*schema.DataMember) []*SchemaMember {
	members := make([]*SchemaMember, 0, len(values))
	for _, value := range values {
		members = append(members, &SchemaMember{
			SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
			Name:           value.Name, Example: value.Example, Sensitive: value.Sensitive, Type: e.projectType(value.Type),
		})
	}
	return members
}

func (e *_SchemaEncoder) projectActor(value *schema.Actor) *SchemaDeclaration {
	actor := new(SchemaActor{Vias: make([]*SchemaActorVia, 0, len(value.Vias))})
	for _, via := range value.Vias {
		actor.Vias = append(actor.Vias, &SchemaActorVia{Name: via.Name})
	}
	if auth := value.Auth; auth != nil {
		actor.Auth = new(SchemaActorAuth{
			Credential: e.projectDataSchema(auth.Credential), Info: e.projectDataSchema(auth.Info),
			IdentifierField: auth.IdentifierField,
		})
	}
	if value.Permission != nil {
		actor.Permission = new(SchemaActorPermission{})
	}
	return new(SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:            value.Pub, Name: value.Name, Kind: schema.DeclarationTypeActor, SkelName: value.SkelName,
		Actor: actor,
	})
}

func (e *_SchemaEncoder) projectResource(value *schema.Resource) *SchemaDeclaration {
	actions := make([]*SchemaResourceAction, 0, len(value.Actions))
	for _, action := range value.Actions {
		actions = append(actions, &SchemaResourceAction{
			SchemaMetadata: metadata(action.Description, action.Deprecated, action.DeprecatedReason),
			Name:           action.Name, PermissionCode: action.PermissionCode, Checks: e.projectResourceChecks(action.Checks),
		})
	}
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:            value.Pub, Name: value.Name, Kind: schema.DeclarationTypeResource, SkelName: value.SkelName,
		Resource: &SchemaResource{Checks: e.projectResourceChecks(value.Checks), Actions: actions},
	}
}

func (e *_SchemaEncoder) projectResourceChecks(values []*schema.ResourceCheck) []*SchemaResourceCheck {
	checks := make([]*SchemaResourceCheck, 0, len(values))
	for _, value := range values {
		checks = append(checks, &SchemaResourceCheck{
			SchemaMetadata: metadata(value.Method.Description, value.Deprecated, value.DeprecatedReason),
			Name:           value.Name, Arguments: e.projectArguments(value.Method.Arguments),
		})
	}
	return checks
}

func (e *_SchemaEncoder) projectService(value *schema.Service) *SchemaDeclaration {
	methods := make([]*SchemaMethod, 0, len(value.Methods))
	for _, method := range value.Methods {
		methods = append(methods, e.projectMethod(method))
	}
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Pub:            value.Public(), Name: value.Name, Kind: schema.DeclarationTypeService, SkelName: value.SkelName,
		Service: &SchemaService{
			Audiences: e.projectAudiences(value.Audiences),
			Api:       value.Api,
			Ext:       value.Ext,
			Auth:      string(value.NormalizedAuth()),
			Require:   e.projectRequirement(value.Require), Methods: methods,
		},
	}
}

func (e *_SchemaEncoder) projectMethod(value *schema.Method) *SchemaMethod {
	return &SchemaMethod{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:           value.Name, SkelName: value.SkelName, Example: value.Example, Auth: string(value.NormalizedAuth()),
		Require: e.projectRequirement(value.Require), InputDescription: value.InputDescription,
		ArgumentsSensitive: value.ArgumentsSensitive, OutputDescription: value.OutputDescription,
		OutputExample: value.OutputExample, ResultSensitive: value.ResultSensitive,
		Arguments: e.projectArguments(value.Arguments), Result: e.projectType(value.ResultType),
	}
}

func (e *_SchemaEncoder) projectArguments(values []*schema.Argument) []*SchemaArgument {
	arguments := make([]*SchemaArgument, 0, len(values))
	for _, value := range values {
		arguments = append(arguments, &SchemaArgument{
			SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
			Name:           value.Name, Example: value.Example, Sensitive: value.Sensitive, Type: e.projectType(value.Type),
		})
	}
	return arguments
}

func (e *_SchemaEncoder) projectWeb(value *schema.Web) *SchemaDeclaration {
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:           value.Name, Kind: schema.DeclarationTypeWeb, SkelName: value.SkelName,
		Web: &SchemaWeb{Auth: string(value.NormalizedAuth()), Audiences: e.projectAudiences(value.Audiences), MountPath: value.MountPath},
	}
}

func (e *_SchemaEncoder) projectTask(value *schema.Task) *SchemaDeclaration {
	triggers := make([]*SchemaTrigger, 0, len(value.Triggers))
	for _, trigger := range value.Triggers {
		triggers = append(triggers, &SchemaTrigger{
			SchemaMetadata: metadata(trigger.Description, trigger.Deprecated, trigger.DeprecatedReason),
			Name:           trigger.Name, SkelName: trigger.SkelName, InputDescription: trigger.InputDescription,
			ArgumentsSensitive: trigger.ArgumentsSensitive, Arguments: e.projectArguments(trigger.Arguments),
		})
	}
	return &SchemaDeclaration{
		SchemaMetadata: metadata(value.Description, value.Deprecated, value.DeprecatedReason),
		Name:           value.Name, Kind: schema.DeclarationTypeTask, SkelName: value.SkelName,
		Task: &SchemaTask{Triggers: triggers},
	}
}

func (e *_SchemaEncoder) projectAudiences(values []*schema.ActorAudience) []*SchemaAudience {
	audiences := make([]*SchemaAudience, 0, len(values))
	for _, value := range values {
		audiences = append(audiences, &SchemaAudience{Actor: e.domain.ReferenceName(value.Actor), Via: value.Via})
	}
	return audiences
}

func (e *_SchemaEncoder) projectRequirement(value *schema.PermissionRequire) *SchemaRequirement {
	if value == nil {
		return nil
	}
	return e.projectRequirementExpr(value.Expression)
}

func (e *_SchemaEncoder) projectRequirementExpr(value *schema.PermissionExpression) *SchemaRequirement {
	if value == nil {
		return nil
	}
	mode := string(value.Mode)
	if mode == "" && value.Check != nil {
		mode = "reference"
	}
	result := &SchemaRequirement{Mode: mode, Code: value.Code}
	if value.Check != nil {
		arguments := make([]*SchemaRequirementCheckArgument, 0, len(value.Check.Arguments))
		for _, argument := range value.Check.Arguments {
			arguments = append(arguments, &SchemaRequirementCheckArgument{Name: argument.Name, JSONPath: argument.JsonPath, Type: e.projectType(argument.Type)})
		}
		result.Check = &SchemaRequirementCheck{
			Resource: e.domain.ReferenceName(value.Check.ResourceSkelName), Action: value.Check.ActionName,
			Check: value.Check.CheckName, Arguments: arguments,
		}
	}
	for _, child := range value.Children {
		result.Children = append(result.Children, e.projectRequirementExpr(child))
	}
	return result
}

func (e *_SchemaEncoder) projectType(value *schema.Type) *SchemaType {
	if value == nil {
		return nil
	}
	result := &SchemaType{Nullable: value.Nullable}
	switch value.Kind {
	case schema.TypeKindUnresolvedReference:
		result.Kind = "importedReference"
		result.Name = e.domain.TypeReferenceName(value)
	case schema.TypeKindScalar:
		result.Kind = "scalar"
		result.Name = strings.ToLower(value.Scalar.Name())
	case schema.TypeKindList:
		result.Kind = "list"
		result.Element = e.projectType(value.List.Value)
	case schema.TypeKindMap:
		result.Kind = "map"
		result.Key = e.projectType(value.Map.Key)
		result.Value = e.projectType(value.Map.Value)
	case schema.TypeKindEnum:
		result.Kind = "enum"
		result.Name = e.domain.TypeReferenceName(value)
	case schema.TypeKindData:
		result.Kind = "data"
		if value.Data != nil {
			switch value.Data.Kind {
			case schema.DataKindConfig:
				result.Kind = "config"
			case schema.DataKindEvent:
				result.Kind = "event"
			}
		}
		result.Name = e.domain.TypeReferenceName(value)
	case schema.TypeKindTypeParameter:
		result.Kind = "typeParameter"
		if value.TypeParameter != nil {
			result.Name = value.TypeParameter.Name
		}
	default:
		result.Kind = fmt.Sprintf("unknown:%d", value.Kind)
	}
	for _, argument := range value.TypeArguments {
		result.Arguments = append(result.Arguments, e.projectType(argument))
	}
	return result
}

func metadata(description string, deprecated bool, reason string) SchemaMetadata {
	return SchemaMetadata{Description: description, Deprecated: deprecated, DeprecatedReason: reason}
}

type _SchemaEncoder struct{ domain *schema.Domain }
