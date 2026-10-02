package analyzer

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

type _DecoratorMeta struct {
	Description      string
	Example          string
	HasExample       bool
	Sensitive        bool
	Deprecated       bool
	DeprecatedReason string
	hasDeprecated    bool
	examplePos       lexer.Position
}

type _DecoratorContext struct {
	allowDesc       bool
	allowExample    bool
	allowSensitive  bool
	allowDeprecated bool
	ignoreOthers    bool
	requireDesc     bool
}

func parseDecoratorMeta(reporter *_DiagnosticReporter, decorators []*grammar.Decorator, ctx _DecoratorContext) (_DecoratorMeta, bool) {
	meta := _DecoratorMeta{}
	valid := true
	for _, decorator := range decorators {
		if reporter.cancelled() {
			break
		}
		switch decorator.Name.Value {
		case "desc":
			accepted := reporter.check(ctx.allowDesc, "%s unexpected decorator %s", decorator.Name.Pos, "@"+decorator.Name.Value)
			accepted = reporter.checkDuplicate(meta.Description == "", "%s duplicated decorator @desc", decorator.Name.Pos) && accepted
			accepted = reporter.check(decorator.Value != nil, "%s decorator @desc requires a string argument", decorator.Name.Pos) && accepted
			valid = accepted && valid
			if !accepted {
				continue
			}
			description, err := grammar.UnquoteDescriptionString(decorator.Value.Raw)
			if err != nil {
				reporter.reportf("%s invalid decorator @desc: %v", decorator.Name.Pos, err)
				valid = false
				continue
			}
			meta.Description = description
		case "example":
			accepted := reporter.check(ctx.allowExample, "%s unexpected decorator %s", decorator.Name.Pos, "@"+decorator.Name.Value)
			accepted = reporter.checkNotDuplicate(meta.HasExample, "%s duplicated decorator @example", decorator.Name.Pos) && accepted
			accepted = reporter.check(decorator.Value != nil && decorator.Value.Raw != "",
				"%s decorator @example requires a value", decorator.Name.Pos) && accepted
			valid = accepted && valid
			if !accepted {
				continue
			}
			meta.Example = decorator.Value.Raw
			meta.HasExample = true
			meta.examplePos = decorator.Name.Pos
		case "sensitive":
			accepted := reporter.check(ctx.allowSensitive, "%s unexpected decorator %s", decorator.Name.Pos, "@"+decorator.Name.Value)
			accepted = reporter.checkNotDuplicate(meta.Sensitive, "%s duplicated decorator @sensitive", decorator.Name.Pos) && accepted
			accepted = reporter.check(decorator.Value == nil,
				"%s decorator @sensitive does not accept an argument", decorator.Name.Pos) && accepted
			valid = accepted && valid
			if accepted {
				meta.Sensitive = true
			}
		case "deprecated":
			accepted := reporter.check(ctx.allowDeprecated, "%s unexpected decorator %s", decorator.Name.Pos, "@"+decorator.Name.Value)
			accepted = reporter.checkNotDuplicate(meta.hasDeprecated, "%s duplicated decorator @deprecated", decorator.Name.Pos) && accepted
			meta.hasDeprecated = true
			accepted = reporter.check(decorator.Value != nil,
				"%s decorator @deprecated requires a deprecation reason", decorator.Name.Pos) && accepted
			valid = accepted && valid
			if !accepted {
				continue
			}
			reason, err := grammar.UnquoteDescriptionString(decorator.Value.Raw)
			if err != nil {
				reporter.reportf("%s invalid decorator @deprecated: %v", decorator.Name.Pos, err)
				valid = false
				continue
			}
			reason = strings.TrimSpace(reason)
			if reason == "" {
				reporter.reportf("%s decorator @deprecated requires a deprecation reason", decorator.Name.Pos)
				valid = false
				continue
			}
			meta.Deprecated = true
			meta.DeprecatedReason = reason
		default:
			valid = reporter.check(ctx.ignoreOthers,
				"%s unexpected decorator %s", decorator.Name.Pos, "@"+decorator.Name.Value) && valid
		}
	}
	valid = reporter.checkNot(ctx.requireDesc && meta.HasExample && meta.Description == "",
		"%s decorator @example must be used with @desc", meta.examplePos) && valid
	return meta, valid
}
