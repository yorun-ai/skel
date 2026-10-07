package source

type _ClientMethodNames struct {
	ReceiverName string
	ContextName  string
	ResultName   string
	ErrorName    string
	OptionsName  string
}

func buildClientMethodNames(arguments []*MethodArgument) *_ClientMethodNames {
	used := make(map[string]bool, len(arguments))
	for _, argument := range arguments {
		used[argument.Name] = true
	}
	allocate := func(name string) string {
		for used[name] {
			name += "_"
		}
		used[name] = true
		return name
	}
	return &_ClientMethodNames{
		ReceiverName: allocate("client"),
		ContextName:  allocate("ctx"),
		ResultName:   allocate("ret"),
		ErrorName:    allocate("err"),
		OptionsName:  allocate("_ivOpts"),
	}
}
