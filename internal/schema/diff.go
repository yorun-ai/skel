package schema

import (
	"fmt"
	"slices"
	"strings"

	"go.yorun.ai/skelc/internal/model"
)

func Diff(baseline, candidate *Document) (*Report, error) {
	if err := Validate(baseline); err != nil {
		return nil, fmt.Errorf("baseline: %w", err)
	}
	if err := Validate(candidate); err != nil {
		return nil, fmt.Errorf("candidate: %w", err)
	}
	diff := &_Diff{usage: collectDiffUsage(baseline, candidate), report: &Report{
		Compatible: true, BaselineDomain: baseline.Domain, CandidateDomain: candidate.Domain,
		Changes: []*Change{},
	}}
	if baseline.Domain != candidate.Domain {
		diff.add(ImpactBreaking, "domain.name.changed", candidate.Domain,
			fmt.Sprintf("domain name changed from %s to %s", baseline.Domain, candidate.Domain), model.Position{}, model.Position{})
		// A domain name identifies the schema root. Replacing it subsumes every
		// nested declaration change, so the report intentionally stops here.
		diff.finish()
		return diff.report, nil
	}
	if baseline.Description != candidate.Description {
		diff.add(ImpactCompatible, "domain.description.changed", candidate.Domain,
			"domain description changed", model.Position{}, model.Position{})
	}

	candidateByName := declarationsByKey(candidate.Declarations)
	matchedBaseline := map[*Declaration]bool{}
	matchedCandidate := map[*Declaration]bool{}
	for _, declaration := range baseline.Declarations {
		other := candidateByName[declarationKey(declaration)]
		if other == nil {
			continue
		}
		diff.compareDeclaration(declaration, other)
		matchedBaseline[declaration] = true
		matchedCandidate[other] = true
	}
	baselineUnmatched := unmatchedDeclarationsByName(baseline.Declarations, matchedBaseline)
	candidateUnmatched := unmatchedDeclarationsByName(candidate.Declarations, matchedCandidate)
	for skelName, baselineValues := range baselineUnmatched {
		candidateValues := candidateUnmatched[skelName]
		if len(baselineValues) != 1 || len(candidateValues) != 1 {
			continue
		}
		diff.compareDeclaration(baselineValues[0], candidateValues[0])
		matchedBaseline[baselineValues[0]] = true
		matchedCandidate[candidateValues[0]] = true
	}
	for _, declaration := range baseline.Declarations {
		if matchedBaseline[declaration] {
			continue
		}
		diff.add(ImpactBreaking, "declaration.removed", declaration.SkelName,
			fmt.Sprintf("%s %s was removed", declaration.Kind, declaration.SkelName), declaration.Pos, model.Position{})
	}
	for _, declaration := range candidate.Declarations {
		if matchedCandidate[declaration] {
			continue
		}
		diff.add(ImpactCompatible, "declaration.added", declaration.SkelName,
			fmt.Sprintf("%s %s was added", declaration.Kind, declaration.SkelName), model.Position{}, declaration.Pos)
	}
	diff.finish()
	return diff.report, nil
}

type _Diff struct {
	usage  map[string]_Usage
	report *Report
}

func (c *_Diff) add(impact ImpactLevel, code, symbol, message string, baseline, candidate model.Position) {
	c.report.Changes = append(c.report.Changes, &Change{
		Code: code, Change: changeType(code), Impact: impact, Symbol: symbol, Message: message,
		Baseline: positionPointer(baseline), Candidate: positionPointer(candidate),
	})
}

func changeType(code string) ChangeType {
	switch {
	case strings.HasSuffix(code, ".added"):
		return ChangeAdded
	case strings.HasSuffix(code, ".removed"):
		return ChangeRemoved
	default:
		return ChangeModified
	}
}

func (c *_Diff) finish() {
	slices.SortFunc(c.report.Changes, func(left, right *Change) int {
		if order := impactOrder(left.Impact) - impactOrder(right.Impact); order != 0 {
			return order
		}
		if order := strings.Compare(left.Symbol, right.Symbol); order != 0 {
			return order
		}
		if order := strings.Compare(left.Code, right.Code); order != 0 {
			return order
		}
		return strings.Compare(left.Message, right.Message)
	})
	for _, change := range c.report.Changes {
		switch change.Impact {
		case ImpactBreaking:
			c.report.Summary.Breaking++
		case ImpactDangerous:
			c.report.Summary.Dangerous++
		case ImpactCompatible:
			c.report.Summary.Compatible++
		}
	}
	c.report.Compatible = c.report.Summary.Breaking == 0
}

func impactOrder(impact ImpactLevel) int {
	switch impact {
	case ImpactBreaking:
		return 1
	case ImpactDangerous:
		return 2
	case ImpactCompatible:
		return 3
	default:
		return 99
	}
}

func positionPointer(position model.Position) *model.Position {
	if position.File == "" && position.Line == 0 && position.Column == 0 {
		return nil
	}
	return new(model.Position(position))
}

func declarationsByKey(values []*Declaration) map[string]*Declaration {
	result := make(map[string]*Declaration, len(values))
	for _, value := range values {
		result[declarationKey(value)] = value
	}
	return result
}

func unmatchedDeclarationsByName(values []*Declaration, matched map[*Declaration]bool) map[string][]*Declaration {
	result := map[string][]*Declaration{}
	for _, value := range values {
		if !matched[value] {
			result[value.SkelName] = append(result[value.SkelName], value)
		}
	}
	return result
}

func declarationKey(value *Declaration) string {
	return string(value.Kind) + "\x00" + value.SkelName
}

func enumItemsByName(values []*EnumItem) map[string]*EnumItem {
	result := make(map[string]*EnumItem, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func membersByName(values []*Member) map[string]*Member {
	result := make(map[string]*Member, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func methodsByName(values []*Method) map[string]*Method {
	result := make(map[string]*Method, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func argumentsByName(values []*Argument) map[string]*Argument {
	result := make(map[string]*Argument, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func resourceActionsByName(values []*ResourceAction) map[string]*ResourceAction {
	result := make(map[string]*ResourceAction, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func resourceChecksByName(values []*ResourceCheck) map[string]*ResourceCheck {
	result := make(map[string]*ResourceCheck, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func triggersByName(values []*Trigger) map[string]*Trigger {
	result := make(map[string]*Trigger, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func audiencesByKey(values []*Audience) map[string]*Audience {
	result := make(map[string]*Audience, len(values))
	for _, value := range values {
		result[value.Actor+"\x00"+value.Via] = value
	}
	return result
}

func actorViaNames(values []*ActorVia) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

func memberNames(values []*Member) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

func argumentNames(values []*Argument) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

func sameNamedSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftSet := stringSet(left)
	for _, value := range right {
		if !leftSet[value] {
			return false
		}
	}
	return true
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func typeDisplay(value *Type) string {
	if value == nil {
		return "void"
	}
	var result string
	switch value.Kind {
	case TypeKindList:
		result = "list<" + typeDisplay(value.Element) + ">"
	case TypeKindMap:
		result = "map<" + typeDisplay(value.Key) + ", " + typeDisplay(value.Value) + ">"
	case TypeKindData, TypeKindConfig, TypeKindEvent, TypeKindImportedReference:
		result = value.Name
		if len(value.Arguments) > 0 {
			arguments := make([]string, 0, len(value.Arguments))
			for _, argument := range value.Arguments {
				arguments = append(arguments, typeDisplay(argument))
			}
			result += "<" + strings.Join(arguments, ", ") + ">"
		}
	case TypeKindEnum, TypeKindScalar, TypeKindTypeParameter:
		result = value.Name
	default:
		result = string(value.Kind)
	}
	if value.Nullable {
		result += "?"
	}
	return result
}
