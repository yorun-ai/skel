package {{ $.PackageName }}{{ template "imports" . }}

func init() { {{ range $task := $.Tasks }}
	task.Register({{ $task.SpecName }}){{ end }}
}
{{ range $task := $.Tasks }}
{{- if $task.CommentLines }}
{{- range $line := $task.CommentLines }}
// {{ $line }}
{{- end }}
{{- else }}
// {{ $task.RunnerName }}
{{- end }}

// {{ $task.Name }} / Spec

var (
	{{ $task.SpecName }} = &task.TaskSpec{
		Name: "{{ $task.Name }}",
		SkelName: "{{ $task.SkelName }}",
		Hash: "{{ $task.Hash }}",

		RunnerType: reflect.TypeFor[{{ $task.RunnerName }}](),
		DefaultRunnerType: reflect.TypeFor[*{{ $task.DefaultRunnerName }}](),
		ERRunnerType: reflect.TypeFor[{{ $task.ERRunnerName }}](),
		WrapperERRunnerCtor: {{ $task.WrapperERRunnerCtorName }},
		DefaultERRunnerType: reflect.TypeFor[*{{ $task.DefaultERRunnerName }}](),
		LauncherType: reflect.TypeFor[{{ $task.LauncherName }}](),
		LauncherCtor: {{ $task.LauncherCtorName }},

		Triggers: []*task.TriggerSpec{ {{ range $task.Triggers }}
			{{ .SpecName }},{{ end }}
		},
	}
{{- range $task.Triggers }}
	{{- if .CommentLines }}
	{{- range $lineIndex, $line := .CommentLines }}
	{{- if eq $lineIndex 0 }}
	// {{ $line }}
	{{- else }}
	//   {{ $line }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{ .SpecName }} = &task.TriggerSpec{
		Name: "{{ .Name }}",
		SkelName: "{{ .SkelName }}",
		LauncherMethodName: "{{ .LaunchName }}",
		RunnerMethodName: "{{ .RunName }}",
		ArgumentsType: {{ if .ArgumentsData }}reflect.TypeFor[{{ .ArgumentsData.Name }}](){{ else }}nil{{ end }},
		ArgumentsSensitive: {{ .ArgumentsSensitive }},
	}{{ end }}
)
{{- if $task.HasTriggerArgs }}

// {{ $task.Name }} / Arguments

{{- range $trigger := $task.Triggers }}
{{- if $trigger.ArgumentsData }}
{{- if $trigger.ArgumentsData.CommentLines }}
{{- range $line := $trigger.ArgumentsData.CommentLines }}
// {{ $line }}
{{- end }}
{{- else }}
// {{ $trigger.ArgumentsData.Name }}
{{- end }}
type {{ $trigger.ArgumentsData.Name }} struct {
{{- range $trigger.ArgumentsData.Members }}
	{{- if .CommentLines }}
	{{- range $line := .CommentLines }}
	// {{ $line }}
	{{- end }}
	{{- end }}
		{{ .Name }} {{ .Type.Plain }} `json:"{{ .SkelName }}"{{ if .Sensitive }} skel:"sensitive"{{ end }}`
{{- end }}
}
{{ end }}
{{ end }}
{{- end }}

// {{ $task.Name }} / Launcher
{{- if $task.DeprecatedCommentLines }}
//
{{- range $task.DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ $task.LauncherName }} interface {
{{- range $task.Triggers }}
	{{- if .LaunchComments }}
	{{- range $lineIndex, $line := .LaunchComments }}
	{{- if eq $lineIndex 0 }}
	// {{ $line }}
	{{- else }}
	//   {{ $line }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{ .LaunchName }}(
	{{- range .Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end -}}
	_ltOpts ...task.LaunchOption)
{{- end }}
}

type {{ $task.LauncherImplName }} struct {
	launcher *task.Launcher
}

func {{ $task.LauncherCtorName }}(launcher *task.Launcher) {{ $task.LauncherName }} {
	return &{{ $task.LauncherImplName }}{
		launcher: launcher,
	}
}
{{ range $trigger := $task.Triggers }}
func (launcher *{{ $task.LauncherImplName }}) {{ $trigger.LaunchName }}(
{{- range $trigger.Arguments }}{{ .Name }} {{ .Type.Plain }}, {{ end -}}
_ltOpts ...task.LaunchOption) {
	launcher.launcher.Launch({{ $trigger.SpecName }}.Info(), {{ if $trigger.ArgumentsData }}&{{ $trigger.ArgumentsData.Name }}{
	{{- range $trigger.Arguments }}
		{{ .MemberName }}: {{ .Name }},
	{{- end }}
	}{{ else }}nil{{ end }}, _ltOpts...)
}
{{ end }}

// {{ $task.Name }} / Runner
{{- if $task.DeprecatedCommentLines }}
//
{{- range $task.DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ $task.RunnerName }} interface {
{{- range $task.Triggers }}
	{{- if .RunComments }}
	{{- range $lineIndex, $line := .RunComments }}
	{{- if eq $lineIndex 0 }}
	// {{ $line }}
	{{- else }}
	//   {{ $line }}
	{{- end }}
	{{- end }}
	{{- end }}
	{{ .RunName }}(
	{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
	)
{{- end }}

	mustBe{{ $task.RunnerName }}()
}

type {{ $task.DefaultRunnerName }} struct{}

{{- range $task.Triggers }}
func (*{{ $task.DefaultRunnerName }}) {{ .RunName }}(
{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
) {
	ex.PanicNew(ex.InvalidRequest, "trigger {{ .SkelName }} is not implemented")
}

{{- end }}
func (*{{ $task.DefaultRunnerName }}) mustBe{{ $task.RunnerName }}() {}

// {{ $task.Name }} / RunnerER
{{- if $task.DeprecatedCommentLines }}
//
{{- range $task.DeprecatedCommentLines }}
// {{ . }}
{{- end }}
{{- end }}

type {{ $task.ERRunnerName }} interface {
{{- range $task.Triggers }}
	{{ .RunName }}(
	{{- range $argIndex, $argument := .Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
	) ex.Error
{{- end }}

	mustBe{{ $task.ERRunnerName }}()
}

type {{ $task.WrapperERRunnerName }} struct {
	{{ $task.DefaultRunnerName }}
	runnerImpl {{ $task.RunnerName }}
}

func {{ $task.WrapperERRunnerCtorName }}(runnerImpl {{ $task.RunnerName }}) {{ $task.ERRunnerName }} {
	return &{{ $task.WrapperERRunnerName }}{
		runnerImpl: runnerImpl,
	}
}

func (runner *{{ $task.WrapperERRunnerName }}) runner() {{ $task.RunnerName }} {
	if runner.runnerImpl == nil {
		return &runner.{{ $task.DefaultRunnerName }}
	}
	return runner.runnerImpl
}
{{ range $trigger := $task.Triggers }}
func (runner *{{ $task.WrapperERRunnerName }}) {{ $trigger.RunName }}(
{{- range $argIndex, $argument := $trigger.Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }} {{ $argument.Type.Plain }}{{ end -}}
) (err ex.Error) {
	defer func() { err = ex.Recover(recover()) }()
	runner.runner().{{ $trigger.RunName }}({{ range $argIndex, $argument := $trigger.Arguments }}{{ if gt $argIndex 0 }}, {{ end }}{{ $argument.Name }}{{ end }})
	return
}

{{ end -}}
func (*{{ $task.WrapperERRunnerName }}) mustBe{{ $task.ERRunnerName }}() {}

type {{ $task.DefaultERRunnerName }} struct {
	{{ $task.WrapperERRunnerName }}
}
{{ end }}
