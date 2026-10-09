package source

import (
	"go/scanner"
	"go/token"
)

// escapeArgumentNames keeps target-language names separate from contract names.
// Matching is case-sensitive: Type is a valid identifier, unlike type.
func escapeArgumentNames(arguments []*MethodArgument, result *Type) {
	reserved := methodReservedNames(arguments, result)
	used := make(map[string]bool, len(arguments))
	for _, argument := range arguments {
		used[argument.Name] = true
	}
	for _, argument := range arguments {
		name := argument.Name
		if !token.Lookup(name).IsKeyword() && !reserved[name] {
			continue
		}
		for {
			name += "_"
			if !used[name] && !reserved[name] {
				break
			}
		}
		argument.Name = name
		used[name] = true
	}
}

func methodReservedNames(arguments []*MethodArgument, result *Type) map[string]bool {
	reserved := map[string]bool{"ex": true, "recover": true, "skeltype": true}
	reserveType := func(kind *Type) {
		if kind == nil {
			return
		}
		// Plain is a generated Go type expression. Its identifiers must remain
		// available when the function body uses a type argument or literal.
		var tokens scanner.Scanner
		file := token.NewFileSet().AddFile("", -1, len(kind.Plain))
		tokens.Init(file, []byte(kind.Plain), nil, 0)
		for {
			_, kind, name := tokens.Scan()
			if kind == token.EOF {
				break
			}
			if kind == token.IDENT {
				reserved[name] = true
			}
		}
	}
	reserveType(result)
	for _, argument := range arguments {
		reserveType(argument.Type)
	}
	return reserved
}

type _MethodNames struct {
	ReceiverName         string
	ContextName          string
	ResultName           string
	ErrorName            string
	OptionsName          string
	ServerReceiverName   string
	LauncherReceiverName string
	RunnerReceiverName   string
}

func buildMethodNames(arguments []*MethodArgument, result *Type) *_MethodNames {
	used := methodReservedNames(arguments, result)
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
	return &_MethodNames{
		ReceiverName:         allocate("client"),
		ContextName:          allocate("ctx"),
		ResultName:           allocate("ret"),
		ErrorName:            allocate("err"),
		OptionsName:          allocate("_ivOpts"),
		ServerReceiverName:   allocate("service"),
		LauncherReceiverName: allocate("launcher"),
		RunnerReceiverName:   allocate("runner"),
	}
}
