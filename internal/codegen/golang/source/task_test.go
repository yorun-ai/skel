package source

import (
	"reflect"
	"testing"

	"go.yorun.ai/skel/internal/model"
)

func TestCastTask(t *testing.T) {
	task_ := new(_Gen).castTask(&model.Task{
		Name:        "RebuildUserIndexTask",
		SkelName:    "demo.user.RebuildUserIndexTask",
		Description: "Rebuild the user index",
		Triggers: []*model.TaskTrigger{
			{
				Name:               "atTime",
				Description:        "Scheduled trigger",
				ArgumentsSensitive: true,
				Arguments: []*model.Argument{
					{
						Name:        "startAt",
						Description: "Start time",
						Example:     `"2026-05-04T12:00:00"`,
						Type: &model.Type{
							Kind:   model.TypeKindScalar,
							Scalar: model.ScalarLocalDateTime,
						},
					},
				},
				ArgumentsData: &model.Data{
					Name: "RebuildUserIndexTaskAtTimeArguments",
					Members: []*model.DataMember{
						{
							Name: "startAt",
							Type: &model.Type{
								Kind:   model.TypeKindScalar,
								Scalar: model.ScalarLocalDateTime,
							},
						},
					},
				},
			},
			{
				Name:        "forGroup",
				Description: "Trigger by user group",
			},
		},
	})

	if len(task_.CommentLines) == 0 || task_.CommentLines[0] != "RebuildUserIndexTaskRunner Rebuild the user index" {
		t.Fatalf("unexpected task comment lines: %+v", task_.CommentLines)
	}
	if !task_.HasTriggerArgs {
		t.Fatal("expected task to contain trigger arguments")
	}
	if len(task_.Triggers) != 2 {
		t.Fatalf("unexpected trigger count: %d", len(task_.Triggers))
	}
	if got := task_.Triggers[0].CommentLines; !reflect.DeepEqual(got, []string{
		"AtTime Scheduled trigger.",
		`@param startAt - Start time (e.g. "2026-05-04T12:00:00")`,
	}) {
		t.Fatalf("unexpected trigger comment lines: %+v", got)
	}
	if got := task_.Triggers[0].LaunchComments; !reflect.DeepEqual(got, []string{
		"LaunchAtTime Scheduled trigger.",
		`@param startAt - Start time (e.g. "2026-05-04T12:00:00")`,
	}) {
		t.Fatalf("unexpected launch comment lines: %+v", got)
	}
	if got := task_.Triggers[0].RunComments; !reflect.DeepEqual(got, []string{
		"RunAtTime Scheduled trigger.",
		`@param startAt - Start time (e.g. "2026-05-04T12:00:00")`,
	}) {
		t.Fatalf("unexpected run comment lines: %+v", got)
	}
	if task_.Triggers[0].LaunchName != "LaunchAtTime" {
		t.Fatalf("unexpected launch name: %s", task_.Triggers[0].LaunchName)
	}
	if task_.Triggers[0].RunName != "RunAtTime" {
		t.Fatalf("unexpected run name: %s", task_.Triggers[0].RunName)
	}
	if !task_.Triggers[0].ArgumentsSensitive {
		t.Fatal("expected trigger arguments to be sensitive")
	}
	if task_.Triggers[0].ArgumentsData == nil || task_.Triggers[0].ArgumentsData.Name != "_RebuildUserIndexTaskAtTimeArguments" {
		t.Fatalf("unexpected arguments data: %+v", task_.Triggers[0].ArgumentsData)
	}
	if task_.Triggers[0].Arguments[0].MemberName != "StartAt" {
		t.Fatalf("unexpected member name: %s", task_.Triggers[0].Arguments[0].MemberName)
	}
}

func TestBuildTaskImports(t *testing.T) {
	imports := buildTaskImports([]*Task{
		{
			Triggers: []*TaskTrigger{
				{
					Arguments: []*MethodArgument{
						{
							Type: &Type{
								Imports: []*Import{{Path: skelImport}},
							},
						},
					},
				},
			},
		},
	})

	if got, want := importPaths(imports), []string{"go.yorun.ai/vine/core/ex", skelImport, "go.yorun.ai/vine/core/task", "reflect"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected import paths: got=%v want=%v", got, want)
	}
}
