package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

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
