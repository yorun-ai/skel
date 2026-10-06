package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skel/internal/codegen/common"
)

func goDocLines(name string, description string) []string {
	lines := common.SplitDocLines(description)
	if len(lines) == 0 {
		return nil
	}

	docLines := make([]string, 0, len(lines))
	docLines = append(docLines, fmt.Sprintf("%s %s", name, lines[0]))
	docLines = append(docLines, lines[1:]...)
	return docLines
}

func deprecatedGoDocLines(lines []string, name string, reason string) []string {
	deprecatedLines := deprecatedGoDocParagraph(reason)
	if len(deprecatedLines) == 0 {
		return lines
	}
	if len(lines) == 0 {
		lines = []string{name}
	}
	lines = append(lines, "")
	return append(lines, deprecatedLines...)
}

func deprecatedGoDocParagraph(reason string) []string {
	reasonLines := common.SplitDocLines(strings.TrimSpace(reason))
	if len(reasonLines) == 0 {
		return nil
	}
	reasonLines[0] = "Deprecated: " + reasonLines[0]
	last := len(reasonLines) - 1
	reasonLines[last] = common.EnsureSentence(reasonLines[last])
	return reasonLines
}

func goMethodDocLines(name string, description string, example string, arguments []*MethodArgument, resultType *Type, outputDescription string, outputExample string, deprecatedReason string) []string {
	docLines := goDocLines(name, common.MergeDescriptionAndExample(description, example))
	if len(docLines) == 0 {
		docLines = []string{name}
	}
	docLines[0] = common.EnsureSentence(docLines[0])

	paramLines := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		if argument.Description == "" {
			continue
		}
		paramLines = append(paramLines, fmt.Sprintf("@param %s - %s", argument.Name, argument.Description))
	}
	if len(paramLines) > 0 {
		docLines = append(docLines, paramLines...)
	}

	outputLines := common.SplitDocLines(common.MergeDescriptionAndExample(outputDescription, outputExample))
	if len(outputLines) > 0 && resultType != nil {
		docLines = append(docLines, fmt.Sprintf("@returns %s - %s", resultType.Plain, outputLines[0]))
	}
	return deprecatedGoDocLines(docLines, name, deprecatedReason)
}
