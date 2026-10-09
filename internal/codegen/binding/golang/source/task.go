package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
	"go.yorun.ai/skel/schema"
)

const taskGoFilename = "task.go"

var taskImports = []*Import{
	{Path: "reflect"},
	{Path: "go.yorun.ai/vine/core/ex"},
	{Path: "go.yorun.ai/vine/core/task"},
}

var taskGoTemplate = loadGoTemplate("task.go.tpl")

type TaskGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Tasks         []*Task
}

type Task struct {
	Name                    string
	SkelName                string
	Hash                    string
	CommentLines            []string
	DeprecatedCommentLines  []string
	SpecName                string
	LauncherName            string
	LauncherImplName        string
	LauncherCtorName        string
	RunnerName              string
	DefaultRunnerName       string
	ERRunnerName            string
	WrapperERRunnerName     string
	WrapperERRunnerCtorName string
	DefaultERRunnerName     string
	Triggers                []*TaskTrigger
	HasTriggerArgs          bool
}

type TaskTrigger struct {
	*_MethodNames
	Name               string
	LaunchName         string
	RunName            string
	SkelName           string
	SpecName           string
	CommentLines       []string
	LaunchComments     []string
	RunComments        []string
	ArgumentsSensitive bool
	Arguments          []*MethodArgument
	ArgumentsData      *Data
}

func (g *_Gen) genTaskGo() {
	payload := g.buildTaskGoPayload()
	if len(payload.Tasks) == 0 {
		return
	}

	g.renderGo(taskGoFilename, taskGoTemplate, payload)
}

func (g *_Gen) buildTaskGoPayload() *TaskGoPayload {
	payload := &TaskGoPayload{
		PackageName: g.pkgName,
		Tasks:       make([]*Task, 0, len(g.view.Tasks)),
	}
	for _, tokenTask := range g.view.Tasks {
		castedTask := g.castTask(tokenTask)
		payload.Tasks = append(payload.Tasks, castedTask)
	}

	imports := buildTaskImports(payload.Tasks)
	payload.StdImports, payload.ModuleImports = splitImports(imports)
	return payload
}

func (g *_Gen) castTask(p *schema.Task) *Task {
	taskName := nameutil.ToCamel(p.Name)
	task_ := &Task{
		Name:     taskName,
		SkelName: p.SkelName,
		Hash:     p.Hash,
		CommentLines: deprecatedGoDocLines(
			goDocLines(fmt.Sprintf("%sRunner", taskName), p.Description),
			fmt.Sprintf("%sRunner", taskName),
			p.DeprecatedReason,
		),
		DeprecatedCommentLines:  deprecatedGoDocParagraph(p.DeprecatedReason),
		SpecName:                fmt.Sprintf("_%sSpec", taskName),
		LauncherName:            fmt.Sprintf("%sLauncher", taskName),
		LauncherImplName:        fmt.Sprintf("_%sLauncher", taskName),
		LauncherCtorName:        fmt.Sprintf("New%sLauncher", taskName),
		RunnerName:              fmt.Sprintf("%sRunner", taskName),
		DefaultRunnerName:       fmt.Sprintf("Default%sRunner", taskName),
		ERRunnerName:            fmt.Sprintf("%sRunnerER", taskName),
		WrapperERRunnerName:     fmt.Sprintf("_Wrapper%sRunnerER", taskName),
		WrapperERRunnerCtorName: fmt.Sprintf("_NewWrapper%sRunnerER", taskName),
		DefaultERRunnerName:     fmt.Sprintf("Default%sRunnerER", taskName),
		Triggers:                make([]*TaskTrigger, 0, len(p.Triggers)),
	}
	for _, trigger := range p.Triggers {
		castedTrigger := g.types.castTaskTrigger(p, trigger)
		task_.Triggers = append(task_.Triggers, castedTrigger)
		if castedTrigger.ArgumentsData != nil {
			task_.HasTriggerArgs = true
		}
	}
	return task_
}

func (r _Types) castTaskTrigger(task_ *schema.Task, p *schema.TaskTrigger) *TaskTrigger {
	arguments := make([]*MethodArgument, 0, len(p.Arguments))
	for _, argument := range p.Arguments {
		castedArgument := r.castMethodArgument(argument)
		arguments = append(arguments, castedArgument)
	}

	escapeArgumentNames(arguments, nil)
	trigger := &TaskTrigger{
		_MethodNames:       buildMethodNames(arguments, nil),
		Name:               nameutil.ToCamel(p.Name),
		LaunchName:         fmt.Sprintf("Launch%s", nameutil.ToCamel(p.Name)),
		RunName:            fmt.Sprintf("Run%s", nameutil.ToCamel(p.Name)),
		SkelName:           p.Name,
		SpecName:           fmt.Sprintf("_%s%sTriggerSpec", task_.Name, nameutil.ToCamel(p.Name)),
		CommentLines:       goMethodDocLines(nameutil.ToCamel(p.Name), p.Description, "", arguments, nil, "", "", p.DeprecatedReason),
		LaunchComments:     goMethodDocLines(fmt.Sprintf("Launch%s", nameutil.ToCamel(p.Name)), p.Description, "", arguments, nil, "", "", p.DeprecatedReason),
		RunComments:        goMethodDocLines(fmt.Sprintf("Run%s", nameutil.ToCamel(p.Name)), p.Description, "", arguments, nil, "", "", p.DeprecatedReason),
		ArgumentsSensitive: p.ArgumentsSensitive,
		Arguments:          arguments,
	}
	if p.ArgumentsData != nil {
		trigger.ArgumentsData = r.castData(p.ArgumentsData)
		trigger.ArgumentsData.Name = fmt.Sprintf("_%s", trigger.ArgumentsData.Name)
		for _, arg := range trigger.Arguments {
			member, ok := sliceutil.Find(trigger.ArgumentsData.Members, func(mem *DataMember) bool {
				return mem.SkelName == arg.SkelName
			})
			if ok {
				arg.MemberName = member.Name
			}
		}
	}
	return trigger
}

func buildTaskImports(tasks []*Task) []*Import {
	imports := newImportSet()
	imports.addMany(taskImports)
	for _, task_ := range tasks {
		for _, trigger := range task_.Triggers {
			for _, argument := range trigger.Arguments {
				imports.addMany(collectTypeImports(argument.Type))
			}
		}
	}
	return imports.sortedValues()
}
