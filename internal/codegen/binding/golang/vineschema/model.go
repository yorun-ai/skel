package vineschema

// The schema model is local template data. Generated Go code still targets
// go.yorun.ai/vine/core/skel, but the compiler does not need to link that
// runtime package merely to render compatible declarations.
type _DomainSchema struct {
	Domain      string
	Description string
	Hash        string
	Full        bool
	Generated   *_GeneratedInfo
	Enums       []*_EnumSchema
	Data        []*_DataSchema
	Configs     []*_ConfigSchema
	Webs        []*_WebSchema
	Events      []*_EventSchema
	Actors      []*_ActorSchema
	Resources   []*_ResourceSchema
	Services    []*_ServiceSchema
	Tasks       []*_TaskSchema
}

type _GeneratedInfo struct {
	CompilerVersion string
}
