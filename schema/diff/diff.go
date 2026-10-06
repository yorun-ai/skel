package diff

import (
	"fmt"
	"slices"
	"strings"

	"go.yorun.ai/skel/schema"
)

// Compare compares analyzed semantic domains without loading sources or imports.
// References are matched by fully qualified identity, whether resolved or not;
// imported declarations are not traversed. Source positions and derived runtime
// data do not contribute to equality. Neither input is modified.
func Compare(baseline, candidate *schema.Domain) (*Report, error) {
	if baseline == nil || candidate == nil {
		return nil, fmt.Errorf("schema comparison requires two non-nil domains")
	}

	diff := &_Diff{baseline: baseline, candidate: candidate, usage: collectDiffUsage(baseline, candidate), report: &Report{
		Compatible: true, BaselineDomain: baseline.Name(), CandidateDomain: candidate.Name(),
		Changes: []*Change{},
	}}
	if baseline.Name() != candidate.Name() {
		diff.add(ImpactBreaking, "domain.name.changed", candidate.Name(),
			fmt.Sprintf("domain name changed from %s to %s", baseline.Name(), candidate.Name()), schema.Position{}, schema.Position{})
		// A domain name identifies the schema root. Replacing it subsumes every
		// nested declaration change, so the report intentionally stops here.
		diff.finish()
		return diff.report, nil
	}
	if baseline.Description() != candidate.Description() {
		diff.add(ImpactCompatible, "domain.description.changed", candidate.Name(),
			"domain description changed", schema.Position{}, schema.Position{})
	}

	baselineDeclarations, candidateDeclarations := baseline.Declarations(), candidate.Declarations()
	candidateByName := declarationsByKey(candidateDeclarations)
	matchedBaseline := map[*schema.Declaration]bool{}
	matchedCandidate := map[*schema.Declaration]bool{}
	for _, declaration := range baselineDeclarations {
		other := candidateByName[declarationKey(declaration)]
		if other == nil {
			continue
		}
		diff.compareDeclaration(declaration, other)
		matchedBaseline[declaration] = true
		matchedCandidate[other] = true
	}
	baselineUnmatched := unmatchedDeclarationsByName(baselineDeclarations, matchedBaseline)
	candidateUnmatched := unmatchedDeclarationsByName(candidateDeclarations, matchedCandidate)
	for skelName, baselineValues := range baselineUnmatched {
		candidateValues := candidateUnmatched[skelName]
		if len(baselineValues) != 1 || len(candidateValues) != 1 {
			continue
		}
		diff.compareDeclaration(baselineValues[0], candidateValues[0])
		matchedBaseline[baselineValues[0]] = true
		matchedCandidate[candidateValues[0]] = true
	}
	for _, declaration := range baselineDeclarations {
		if matchedBaseline[declaration] {
			continue
		}
		diff.add(ImpactBreaking, "declaration.removed", declaration.SkelName,
			fmt.Sprintf("%s %s was removed", declaration.Kind, declaration.SkelName), declaration.Pos, schema.Position{})
	}
	for _, declaration := range candidateDeclarations {
		if matchedCandidate[declaration] {
			continue
		}
		diff.add(ImpactCompatible, "declaration.added", declaration.SkelName,
			fmt.Sprintf("%s %s was added", declaration.Kind, declaration.SkelName), schema.Position{}, declaration.Pos)
	}
	diff.finish()
	return diff.report, nil
}

type _Diff struct {
	baseline, candidate *schema.Domain
	usage               map[string]_Usage
	report              *Report
}

func (c *_Diff) add(impact ImpactLevel, code, symbol, message string, baseline, candidate schema.Position) {
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

func positionPointer(position schema.Position) *schema.Position {
	if position.File == "" && position.Line == 0 && position.Column == 0 {
		return nil
	}
	return new(schema.Position(position))
}

func declarationsByKey(values []*schema.Declaration) map[string]*schema.Declaration {
	result := make(map[string]*schema.Declaration, len(values))
	for _, value := range values {
		result[declarationKey(value)] = value
	}
	return result
}

func unmatchedDeclarationsByName(values []*schema.Declaration, matched map[*schema.Declaration]bool) map[string][]*schema.Declaration {
	result := map[string][]*schema.Declaration{}
	for _, value := range values {
		if !matched[value] {
			result[value.SkelName] = append(result[value.SkelName], value)
		}
	}
	return result
}

func declarationKey(value *schema.Declaration) string {
	return string(value.Kind) + "\x00" + value.SkelName
}

func enumItemsByName(values []*schema.EnumItem) map[string]*schema.EnumItem {
	result := make(map[string]*schema.EnumItem, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func membersByName(values []*schema.DataMember) map[string]*schema.DataMember {
	result := make(map[string]*schema.DataMember, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func methodsByName(values []*schema.Method) map[string]*schema.Method {
	result := make(map[string]*schema.Method, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func argumentsByName(values []*schema.Argument) map[string]*schema.Argument {
	result := make(map[string]*schema.Argument, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func resourceActionsByName(values []*schema.ResourceAction) map[string]*schema.ResourceAction {
	result := make(map[string]*schema.ResourceAction, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func resourceChecksByName(values []*schema.ResourceCheck) map[string]*schema.ResourceCheck {
	result := make(map[string]*schema.ResourceCheck, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func triggersByName(values []*schema.TaskTrigger) map[string]*schema.TaskTrigger {
	result := make(map[string]*schema.TaskTrigger, len(values))
	for _, value := range values {
		result[value.Name] = value
	}
	return result
}

func audiencesByKey(domain *schema.Domain, values []*schema.ActorAudience) map[string]*schema.ActorAudience {
	result := make(map[string]*schema.ActorAudience, len(values))
	for _, value := range values {
		result[referenceName(domain, value.Actor)+"\x00"+value.Via] = value
	}
	return result
}

func actorViaNames(values []*schema.ActorVia) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

func memberNames(values []*schema.DataMember) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.Name)
	}
	return result
}

func argumentNames(values []*schema.Argument) []string {
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
