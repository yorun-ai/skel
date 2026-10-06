package skeleton

import (
	"strconv"
	"strings"

	"go.yorun.ai/skel/schema"
)

type _DecoratorView struct {
	Indent    int
	Quoted    string
	Lines     []string
	Multiline bool
	Object    bool
}

func descriptionView(desc string, indent int) *_DecoratorView {
	if desc == "" {
		return nil
	}
	if strings.Contains(desc, "\n") && !strings.Contains(desc, `"""`) {
		return &_DecoratorView{Indent: indent, Lines: strings.Split(desc, "\n"), Multiline: true}
	}
	return &_DecoratorView{Indent: indent, Quoted: strconv.Quote(desc)}
}

func exampleView(example string, indent int) *_DecoratorView {
	if example == "" {
		return nil
	}
	if strings.Contains(example, "\n") {
		lines := strings.Split(strings.TrimSpace(example), "\n")
		if len(lines) >= 2 && strings.HasPrefix(lines[0], "{") && lines[len(lines)-1] == "}" {
			return &_DecoratorView{Indent: indent, Lines: lines[1 : len(lines)-1], Multiline: true, Object: true}
		}
		return &_DecoratorView{Indent: indent, Lines: lines, Multiline: true}
	}
	return &_DecoratorView{Indent: indent, Quoted: example}
}

func sensitiveView(sensitive bool, indent int) *_DecoratorView {
	if !sensitive {
		return nil
	}
	return &_DecoratorView{Indent: indent}
}

func deprecatedView(deprecated bool, reason string, indent int) *_DecoratorView {
	if !deprecated {
		return nil
	}
	return descriptionView(reason, indent)
}

func emptyMethod(method *schema.Method, service *schema.Service) bool {
	return methodAuthMarker(method) == "" && len(method.Arguments) == 0 && method.ResultType == nil
}

func authMarker(mode schema.AuthMode) string {
	if mode == schema.AuthModeAuth || mode == schema.AuthModeNoAuth {
		return string(mode)
	}
	if mode == schema.AuthModeRequired || mode == schema.AuthModeOptional || mode == schema.AuthModeAnonymous || mode == schema.AuthModeOff {
		return "auth " + string(mode)
	}
	return ""
}

func methodAuthMarker(method *schema.Method) string {
	return authMarker(method.AuthMode)
}

func importAlias(import_ *schema.Import) string {
	if !import_.ExplicitAlias {
		return ""
	}
	return import_.Alias
}

func typeParameterNames(params []*schema.TypeParameter) []string {
	names := make([]string, 0, len(params))
	for _, param := range params {
		names = append(names, param.Name)
	}
	return names
}

func configLifecycle(config *schema.Data) string {
	return string(config.Lifecycle)
}
