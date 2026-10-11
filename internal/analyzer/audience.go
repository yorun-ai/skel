package analyzer

import (
	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/diagnostic"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

type _AudienceState struct {
	pos    lexer.Position
	legacy bool
}

func parseActorAudiences(reporter *_DiagnosticReporter, audiences []*grammar.ActorAudience) ([]*schema.ActorAudience, bool) {
	if len(audiences) == 0 {
		return []*schema.ActorAudience{}, true
	}
	parsed := make([]*schema.ActorAudience, 0, len(audiences))
	seen := map[string]_AudienceState{}
	valid := true
	for index, clause := range audiences {
		if index > 0 {
			reporter.warnf(diagnostic.CodeRepeatedFor, "%s repeated for clauses are deprecated; combine actors in one comma-separated for clause", clause.Pos)
		}
		for _, audience := range clause.Actors {
			if reporter.cancelled() {
				break
			}
			actorIdent := audience.Actor.Parts[len(audience.Actor.Parts)-1]
			valid = checkCaseAdvanced(reporter, "Actor", "", "Actor", caseTypeCamel, actorIdent) && valid
			if audience.Via != nil {
				reporter.warnIgnoredVia(audience.Via.Pos)
			}
			if audience.Block != nil {
				reporter.warnf(diagnostic.CodeForBlockIgnored, "%s empty for blocks are deprecated and ignored; remove the braces", audience.Block.Pos)
			}
			name := audience.Actor.String()
			previous, duplicated := seen[name]
			if duplicated {
				// Legacy transport-specific entries collapse to one actor audience.
				if audience.Via != nil || previous.legacy {
					previous.legacy = true
					seen[name] = previous
					continue
				}
				reporter.reportDuplicatef("%s duplicated service audience %s found, also present at %s", actorIdent.Pos, name, previous.pos)
				valid = false
				continue
			}
			seen[name] = _AudienceState{pos: actorIdent.Pos, legacy: audience.Via != nil}
			parsed = append(parsed, new(schema.ActorAudience{
				Actor: name,
				Pos:   position(audience.Pos),
			}))
		}
	}
	return parsed, valid
}
