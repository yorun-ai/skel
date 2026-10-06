package golang_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
)

func writeFileForTest(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create test directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(content)
}

func newModelDomainForTest(t *testing.T, spec model.DomainSpec) *model.Domain {
	t.Helper()

	prepareModelSpecForTest(&spec)
	domain := model.NewDomainFromSpec(spec)
	fillModelHashesForTest(domain)
	return domain
}

func prepareModelSpecForTest(spec *model.DomainSpec) {
	for _, enum := range spec.Enums {
		enum.Domain = spec.Name
		setModelSkelNameForTest(spec.Name, enum.Name, &enum.SkelName)
		if enum.UnspecifiedItem == nil {
			enum.UnspecifiedItem = &model.EnumItem{Name: "UNSPECIFIED"}
		}
	}
	prepareModelDataForTest(spec.Name, spec.Data, model.DataKindData)
	prepareModelDataForTest(spec.Name, spec.Configs, model.DataKindConfig)
	prepareModelDataForTest(spec.Name, spec.Events, model.DataKindEvent)
	for _, actor := range spec.Actors {
		setModelSkelNameForTest(spec.Name, actor.Name, &actor.SkelName)
		if actor.AuthEnabled && actor.AuthService == nil {
			prepareModelDataForTest(spec.Name, []*model.Data{actor.AuthCredential, actor.AuthInfo}, model.DataKindData)
			method := &model.Method{
				Name:       "auth",
				SkelName:   "auth",
				Auth:       model.AuthModeNoAuth,
				ResultType: codegentest.DataType(actor.AuthInfo),
				Arguments: []*model.Argument{
					{Name: "credential", Type: codegentest.DataType(actor.AuthCredential)},
				},
			}
			actor.AuthMethod = method
			actor.AuthService = &model.Service{
				Name:     actor.Name + "AuthService",
				SkelName: spec.Name + "." + actor.Name + "AuthService",
				Auth:     model.AuthModeNoAuth,
				Methods:  []*model.Method{method},
			}
		}
	}
	for _, service := range spec.Services {
		setModelSkelNameForTest(spec.Name, service.Name, &service.SkelName)
		for _, method := range service.Methods {
			if method.SkelName == "" {
				method.SkelName = method.Name
			}
		}
	}
	sort.Slice(spec.Enums, func(i, j int) bool { return spec.Enums[i].Name < spec.Enums[j].Name })
	sort.Slice(spec.Data, func(i, j int) bool { return spec.Data[i].Name < spec.Data[j].Name })
	sort.Slice(spec.Configs, func(i, j int) bool { return spec.Configs[i].Name < spec.Configs[j].Name })
	sort.Slice(spec.Events, func(i, j int) bool { return spec.Events[i].Name < spec.Events[j].Name })
	sort.Slice(spec.Actors, func(i, j int) bool { return spec.Actors[i].Name < spec.Actors[j].Name })
	sort.Slice(spec.Services, func(i, j int) bool { return spec.Services[i].Name < spec.Services[j].Name })
}

func prepareModelDataForTest(domain string, values []*model.Data, kind model.DataKind) {
	for _, value := range values {
		if value == nil {
			continue
		}
		value.Domain = domain
		value.Kind = kind
		setModelSkelNameForTest(domain, value.Name, &value.SkelName)
	}
}

func setModelSkelNameForTest(domain string, name string, target *string) {
	if *target == "" {
		*target = strings.TrimSuffix(domain, ".") + "." + name
	}
}

func methodForTest(serviceName string, method *model.Method) *model.Method {
	if len(method.Arguments) > 0 && method.ArgumentsData == nil {
		method.ArgumentsData = argumentsDataForTest(serviceName+nameutil.ToCamel(method.Name), method.Arguments)
	}
	return method
}

func triggerForTest(taskName string, trigger *model.TaskTrigger) *model.TaskTrigger {
	if len(trigger.Arguments) > 0 && trigger.ArgumentsData == nil {
		trigger.ArgumentsData = argumentsDataForTest(taskName+nameutil.ToCamel(trigger.Name), trigger.Arguments)
	}
	return trigger
}

func argumentsDataForTest(owner string, args []*model.Argument) *model.Data {
	members := make([]*model.DataMember, 0, len(args))
	for _, arg := range args {
		members = append(members, &model.DataMember{
			Name:        arg.Name,
			Description: arg.Description,
			Example:     arg.Example,
			Sensitive:   arg.Sensitive,
			Type:        arg.Type,
		})
	}
	return &model.Data{
		Name:    owner + "Arguments",
		Members: members,
	}
}

func fillModelHashesForTest(domain *model.Domain) {
	model.SetDomainHash(domain, "domain-hash")
	for _, enum := range domain.Enums() {
		enum.Hash = "enum-hash"
	}
	for _, data := range domain.Data() {
		data.Hash = "data-hash"
	}
	for _, event := range domain.Events() {
		event.Hash = "event-hash"
	}
	for _, actor := range domain.Actors() {
		actor.Hash = "actor-hash"
	}
	for _, service := range domain.Services() {
		service.Hash = "service-hash"
		for _, method := range service.Methods {
			method.Hash = "method-hash"
		}
	}
	for _, task := range domain.Tasks() {
		task.Hash = "task-hash"
		for _, trigger := range task.Triggers {
			trigger.Hash = "trigger-hash"
		}
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file %s to be missing, err=%v", path, err)
	}
}
