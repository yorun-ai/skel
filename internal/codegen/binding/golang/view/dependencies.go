package view

import (
	"go.yorun.ai/skel/internal/codegen"
	"go.yorun.ai/skel/schema"
)

func (v *Domain) TypeRoots() []*schema.Type {
	return (codegen.Declarations{Enums: v.Enums, Data: v.Data, Configs: v.Configs, Events: v.Events, Actors: v.Actors, Resources: v.Resources, Webs: v.Webs, Services: v.Services, Tasks: v.Tasks}).TypeRoots(v.mode == ModeApi)
}
