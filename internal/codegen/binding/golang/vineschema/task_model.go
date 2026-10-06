package vineschema

type _TaskSchema struct {
	Name             string
	SkelName         string
	Description      string
	Deprecated       bool
	DeprecatedReason string
	Hash             string
	Triggers         []*_TriggerSchema
}

type _TriggerSchema struct {
	Name               string
	SkelName           string
	Description        string
	Deprecated         bool
	DeprecatedReason   string
	Hash               string
	InputDescription   string
	ArgumentsSensitive bool
	Arguments          []*_MemberSchema
}
