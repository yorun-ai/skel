package golang_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.yorun.ai/skel/internal/codegen/binding/golang"
	"go.yorun.ai/skel/internal/codegen/codegentest"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/schema"
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

func newSchemaDomainForTest(t *testing.T, spec schema.DomainSpec) *schema.Domain {
	t.Helper()

	prepareSchemaSpecForTest(&spec)
	spec.Hash = "domain-hash"
	domain := schema.NewDomainFromSpec(spec)
	fillSchemaHashesForTest(domain)
	return domain
}

func prepareSchemaSpecForTest(spec *schema.DomainSpec) {
	for _, enum := range spec.Enums {
		enum.Domain = spec.Name
		setSchemaSkelNameForTest(spec.Name, enum.Name, &enum.SkelName)
		if enum.UnspecifiedItem == nil {
			enum.UnspecifiedItem = &schema.EnumItem{Name: "UNSPECIFIED"}
		}
	}
	prepareSchemaDataForTest(spec.Name, spec.Data, schema.DataKindData)
	prepareSchemaDataForTest(spec.Name, spec.Configs, schema.DataKindConfig)
	prepareSchemaDataForTest(spec.Name, spec.Events, schema.DataKindEvent)
	for _, actor := range spec.Actors {
		setSchemaSkelNameForTest(spec.Name, actor.Name, &actor.SkelName)
		if actor.Auth != nil && actor.Auth.Service == nil {
			prepareSchemaDataForTest(spec.Name, []*schema.Data{actor.Auth.Credential, actor.Auth.Info}, schema.DataKindData)
			method := &schema.Method{
				Name:       "auth",
				SkelName:   "auth",
				AuthMode:   schema.AuthModeNoAuth,
				ResultType: codegentest.DataType(actor.Auth.Info),
				Arguments: []*schema.Argument{
					{Name: "credential", Type: codegentest.DataType(actor.Auth.Credential)},
				},
			}
			actor.Auth.Method = method
			actor.Auth.Service = &schema.Service{
				Name:     actor.Name + "AuthService",
				SkelName: spec.Name + "." + actor.Name + "AuthService",
				AuthMode: schema.AuthModeNoAuth,
				Methods:  []*schema.Method{method},
			}
		}
	}
	for _, service := range spec.Services {
		setSchemaSkelNameForTest(spec.Name, service.Name, &service.SkelName)
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

func prepareSchemaDataForTest(domain string, values []*schema.Data, kind schema.DataKind) {
	for _, value := range values {
		if value == nil {
			continue
		}
		value.Domain = domain
		value.Kind = kind
		setSchemaSkelNameForTest(domain, value.Name, &value.SkelName)
	}
}

func setSchemaSkelNameForTest(domain string, name string, target *string) {
	if *target == "" {
		*target = strings.TrimSuffix(domain, ".") + "." + name
	}
}

func methodForTest(serviceName string, method *schema.Method) *schema.Method {
	if len(method.Arguments) > 0 && method.ArgumentsData == nil {
		method.ArgumentsData = argumentsDataForTest(serviceName+nameutil.ToCamel(method.Name), method.Arguments)
	}
	return method
}

func triggerForTest(taskName string, trigger *schema.TaskTrigger) *schema.TaskTrigger {
	if len(trigger.Arguments) > 0 && trigger.ArgumentsData == nil {
		trigger.ArgumentsData = argumentsDataForTest(taskName+nameutil.ToCamel(trigger.Name), trigger.Arguments)
	}
	return trigger
}

func argumentsDataForTest(owner string, args []*schema.Argument) *schema.Data {
	members := make([]*schema.DataMember, 0, len(args))
	for _, arg := range args {
		members = append(members, &schema.DataMember{
			Name:        arg.Name,
			Description: arg.Description,
			Example:     arg.Example,
			Sensitive:   arg.Sensitive,
			Type:        arg.Type,
		})
	}
	return &schema.Data{
		Name:    owner + "Arguments",
		Members: members,
	}
}

func fillSchemaHashesForTest(domain *schema.Domain) {
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

func generateFixture(domain *schema.Domain, option golang.Option) error {
	if option.CompilerVersion == "" {
		option.CompilerVersion = "v0.0.0-dev"
	}
	resolved, err := golang.ResolveOption(option)
	if err != nil {
		return err
	}
	return golang.Generate(domain, resolved)
}
