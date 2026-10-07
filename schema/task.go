package schema

// Task describes a background task declaration.
type Task struct {
	// Pos is the task declaration's source position.
	Pos Position
	// Name is the task's local name.
	Name string
	// SkelName is the task's fully qualified Skel name.
	SkelName string
	// Hash is the task's compatibility hash.
	Hash string
	// Description is the task's documentation text.
	Description string
	// Deprecated reports whether the task should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the task is deprecated and what to use instead.
	DeprecatedReason string
	// Triggers lists task triggers in source order.
	Triggers []*TaskTrigger
}

// TaskTrigger describes one way to invoke a task.
type TaskTrigger struct {
	// Pos is the trigger declaration's source position.
	Pos Position
	// Name is the trigger's normalized local name.
	Name string
	// SkelName is the trigger name as represented in Skel metadata.
	SkelName string
	// Hash is the trigger's compatibility hash.
	Hash string
	// Description is the trigger's documentation text.
	Description string
	// Deprecated reports whether the trigger should no longer be used.
	Deprecated bool
	// DeprecatedReason explains why the trigger is deprecated and what to use instead.
	DeprecatedReason string
	// InputDescription documents the trigger input as a whole.
	InputDescription string
	// ArgumentsSensitive reports whether the trigger input is sensitive as a whole.
	ArgumentsSensitive bool
	// Arguments lists trigger arguments in source order.
	Arguments []*Argument
	// ArgumentsData is the language-defined data schema representing trigger arguments.
	ArgumentsData *Data
}
