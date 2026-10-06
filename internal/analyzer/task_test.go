package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseTask(t *testing.T) {
	task := parseTaskTest(t, &grammar.Task{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue(`"Rebuild the user index"`)},
		},
		Name: ident("RebuildUserIndexTask"),
		Triggers: []*grammar.TaskTrigger{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("desc"), Value: decoratorValue(`"Scheduled trigger"`)},
				},
				Name: ident("atTime"),
				Input: &grammar.MethodInput{
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: decoratorValue(`"Trigger parameters"`)},
						{Name: ident("sensitive")},
					},
					Arguments: []*grammar.Argument{
						{
							Decorators: []*grammar.Decorator{
								{Name: ident("desc"), Value: decoratorValue(`"Start time"`)},
							},
							Name: ident("startAt"),
							Type: plainType(grammar.LocalDateTime),
						},
					},
				},
			},
		},
	})
	if task.Name != "RebuildUserIndexTask" {
		t.Fatalf("unexpected task name: %s", task.Name)
	}
	if task.Description != "Rebuild the user index" {
		t.Fatalf("unexpected task description: %q", task.Description)
	}
	if len(task.Triggers) != 1 || task.Triggers[0].Name != "atTime" {
		t.Fatalf("unexpected task triggers: %+v", task.Triggers)
	}
	if task.Triggers[0].InputDescription != "Trigger parameters" {
		t.Fatalf("unexpected input description: %q", task.Triggers[0].InputDescription)
	}
	if !task.Triggers[0].ArgumentsSensitive {
		t.Fatal("expected trigger input to be sensitive")
	}
	if task.Triggers[0].ArgumentsData == nil || task.Triggers[0].ArgumentsData.Name != "RebuildUserIndexTaskAtTimeArguments" {
		t.Fatalf("unexpected arguments data: %+v", task.Triggers[0].ArgumentsData)
	}
}

func TestParseTaskTriggerWithoutInput(t *testing.T) {
	task := parseTaskTest(t, &grammar.Task{
		Name: ident("RebuildUserIndexTask"),
		Triggers: []*grammar.TaskTrigger{
			{Name: ident("atTime")},
		},
	})
	if len(task.Triggers) != 1 || task.Triggers[0].Name != "atTime" {
		t.Fatalf("unexpected task triggers: %+v", task.Triggers)
	}
	if len(task.Triggers[0].Arguments) != 0 {
		t.Fatalf("unexpected trigger arguments: %+v", task.Triggers[0].Arguments)
	}
	if task.Triggers[0].ArgumentsData != nil {
		t.Fatalf("unexpected arguments data: %+v", task.Triggers[0].ArgumentsData)
	}
}

func TestParseTaskRejectsTriggerExampleDecorator(t *testing.T) {
	expectTaskDiagnostic(t, "unexpected decorator @example", &grammar.Task{
		Name: ident("RebuildUserIndexTask"),
		Triggers: []*grammar.TaskTrigger{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("example"), Value: decoratorValue(`"demo"`)},
				},
				Name: ident("atTime"),
				Input: &grammar.MethodInput{
					Arguments: []*grammar.Argument{
						{Name: ident("startAt"), Type: plainType(grammar.LocalDateTime)},
					},
				},
			},
		},
	})
}

func TestParseTaskRejectsInputExampleDecorator(t *testing.T) {
	expectTaskDiagnostic(t, "unexpected decorator @example", &grammar.Task{
		Name: ident("RebuildUserIndexTask"),
		Triggers: []*grammar.TaskTrigger{
			{
				Name: ident("atTime"),
				Input: &grammar.MethodInput{
					Decorators: []*grammar.Decorator{
						{Name: ident("example"), Value: decoratorValue(`{"startAt":"2026-05-01T10:00:00"}`)},
					},
					Arguments: []*grammar.Argument{
						{Name: ident("startAt"), Type: plainType(grammar.LocalDateTime)},
					},
				},
			},
		},
	})
}
