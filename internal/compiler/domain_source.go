package compiler

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/internal/parser/grammar"
)

type _DomainSourceIssue struct {
	code          string
	position      model.Position
	message       string
	strictMessage string
}

func mergeDomainContents(contents []*grammar.SkelContent) *grammar.SkelContent {
	if len(contents) == 0 {
		return new(grammar.SkelContent)
	}
	ordered := append([]*grammar.SkelContent{}, contents...)
	slices.SortFunc(ordered, func(left, right *grammar.SkelContent) int {
		return strings.Compare(left.Pos.Filename, right.Pos.Filename)
	})
	domainContent := ordered[0].Domain
	for _, content := range ordered {
		if filepath.Base(content.Pos.Filename) == loader.DomainFileName {
			domainContent = content.Domain
			break
		}
	}
	merged := &grammar.SkelContent{Domain: domainContent}
	if domainContent != nil {
		merged.Pos = domainContent.Pos
	}
	for _, content := range ordered {
		merged.Imports = append(merged.Imports, content.Imports...)
		merged.Entries = append(merged.Entries, content.Entries...)
	}
	return merged
}

func inspectDirectorySource(path, expectedDomain string, content *grammar.SkelContent) *_DomainSourceIssue {
	position := model.Position{File: path, Line: 1, Column: 1}
	if content != nil {
		position = parser.SourcePosition(content.Pos)
		position.File = path
	}
	if content == nil || content.Domain == nil || content.Domain.Name == nil {
		message := "missing domain declaration"
		return &_DomainSourceIssue{
			code: DiagnosticCodeDomainMissing, position: position,
			message: message, strictMessage: fmt.Sprintf("%s in %s", message, path),
		}
	}
	position = parser.SourcePosition(content.Domain.Name.Pos)
	position.File = path
	name := content.Domain.Name.String()
	if expectedDomain != "" && name != expectedDomain {
		return &_DomainSourceIssue{
			code: DiagnosticCodeDomainMismatch, position: position,
			message:       fmt.Sprintf("domain mismatch: found=%s, expected=%s", name, expectedDomain),
			strictMessage: fmt.Sprintf("domain mismatch in %s: found=%s, expected=%s", path, name, expectedDomain),
		}
	}
	if filepath.Base(path) == loader.DomainFileName {
		if len(content.Entries) != 0 {
			return &_DomainSourceIssue{
				code: DiagnosticCodeDomainFileContent, position: position,
				message:       fmt.Sprintf("%s can only contain domain declaration and @desc", path),
				strictMessage: fmt.Sprintf("%s can only contain domain declaration and @desc", path),
			}
		}
		return nil
	}
	if len(content.Domain.Decorators) != 0 {
		return &_DomainSourceIssue{
			code: DiagnosticCodeDomainDecorator, position: position,
			message:       fmt.Sprintf("domain decorator is only allowed in %s: %s", loader.DomainFileName, path),
			strictMessage: fmt.Sprintf("domain decorator is only allowed in %s: %s", loader.DomainFileName, path),
		}
	}
	return nil
}
