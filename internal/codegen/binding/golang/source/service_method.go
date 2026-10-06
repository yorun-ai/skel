package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

type ServiceMethod struct {
	*_ClientMethodNames
	Name                        string
	SkelName                    string
	SpecName                    string
	CommentLines                []string
	Arguments                   []*MethodArgument
	ArgumentsData               *Data
	ResultType                  *Type
	ArgumentsSensitive          bool
	ResultSensitive             bool
	ArgumentsContainsBinaryType bool
	ResultContainsBinaryType    bool
}

func (r _Types) castServiceMethod(ps *schema.Service, pm *schema.Method) *ServiceMethod {
	methodArgs := make([]*MethodArgument, 0, len(pm.Arguments))
	for _, argument := range pm.Arguments {
		castedArgument := r.castMethodArgument(argument)
		methodArgs = append(methodArgs, castedArgument)
	}
	resultType := r.castType(pm.ResultType)
	method := &ServiceMethod{
		_ClientMethodNames:          buildClientMethodNames(methodArgs),
		Name:                        nameutil.ToCamel(pm.Name),
		SkelName:                    pm.Name,
		Arguments:                   methodArgs,
		ResultType:                  resultType,
		ArgumentsSensitive:          pm.ArgumentsSensitive,
		ResultSensitive:             pm.ResultSensitive,
		ArgumentsContainsBinaryType: methodArgumentsContainBinaryType(pm),
		ResultContainsBinaryType:    methodResultContainsBinaryType(pm),
	}
	method.SpecName = fmt.Sprintf("_%s%sSpec", ps.Name, method.Name)
	if pm.ArgumentsData != nil {
		method.ArgumentsData = r.castData(pm.ArgumentsData)
		method.ArgumentsData.Name = fmt.Sprintf("_%s", method.ArgumentsData.Name)
		for _, arg := range method.Arguments {
			member, ok := sliceutil.Find(method.ArgumentsData.Members, func(mem *DataMember) bool {
				return mem.SkelName == arg.SkelName
			})
			if ok {
				arg.MemberName = member.Name
			}
		}
	}
	method.CommentLines = goMethodDocLines(
		method.Name,
		pm.Description,
		pm.Example,
		method.Arguments,
		method.ResultType,
		pm.OutputDescription,
		pm.OutputExample,
		pm.DeprecatedReason,
	)
	return method
}

func methodArgumentsContainBinaryType(method *schema.Method) bool {
	for _, argument := range method.Arguments {
		if argument.Type.ContainsBinaryType() {
			return true
		}
	}
	return false
}

func methodResultContainsBinaryType(method *schema.Method) bool {
	return method.ResultType.ContainsBinaryType()
}

type MethodArgument struct {
	Name        string
	SkelName    string
	MemberName  string
	Description string
	Type        *Type
}

func (r _Types) castMethodArgument(p *schema.Argument) *MethodArgument {
	argType := r.castType(p.Type)
	name := nameutil.ToLowerCamel(p.Name)
	// Generated method bodies may refer to the scalar package, including when
	// a business parameter is named types. Keep its wire name unchanged.
	if name == "types" {
		name = "types_"
	}
	description := binding.MergeDescriptionAndExample(p.Description, p.Example)
	if p.Deprecated {
		if description != "" {
			description += " "
		}
		description += "Deprecated: " + binding.EnsureSentence(p.DeprecatedReason)
	}
	return &MethodArgument{
		Name:        name,
		SkelName:    p.Name,
		Description: description,
		Type:        argType,
	}
}
