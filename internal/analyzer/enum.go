package analyzer

import (
	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

const unspecifiedEnumName = "UNSPECIFIED"

func parseEnum(reporter *_DiagnosticReporter, ge *grammar.Enum) (*schema.Enum, bool) {
	valid := checkCase(reporter, "Enum", caseTypeCamel, ge.Name)
	valid = checkNotReservedKindSuffix(reporter, "Enum", ge.Name) && valid

	enum := &schema.Enum{
		Pos:  position(ge.Name.Pos),
		Name: ge.Name.Value,
		Pub:  ge.Pub,
		UnspecifiedItem: &schema.EnumItem{
			Name: unspecifiedEnumName,
		},
		Items: []*schema.EnumItem{},
	}
	meta, metaValid := parseDecoratorMeta(reporter, ge.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid
	enum.Description = meta.Description
	enum.Deprecated = meta.Deprecated
	enum.DeprecatedReason = meta.DeprecatedReason
	itemPositionByName := map[string]lexer.Position{}

	for _, grammarItem := range ge.Items {
		if reporter.cancelled() {
			break
		}
		item, itemValid := parseEnumItem(reporter, grammarItem)
		valid = itemValid && valid
		duplicatedPosition, duplicated := itemPositionByName[item.Name]
		if duplicated {
			reporter.reportDuplicatef("%s duplicated EnumItem %s found, also present at %s", item.Pos, item.Name, duplicatedPosition)
			valid = false
			continue
		}
		itemPositionByName[item.Name] = lexer.Position{Filename: item.Pos.File, Line: item.Pos.Line, Column: item.Pos.Column}
		enum.Items = append(enum.Items, item)
	}

	valid = reporter.check(len(enum.Items) > 0, "%s missing EnumItem for %s", enum.Pos, enum.Name) && valid
	return enum, valid
}

func parseEnumItem(reporter *_DiagnosticReporter, gei *grammar.EnumItem) (*schema.EnumItem, bool) {
	valid := checkCase(reporter, "EnumItem", caseTypeScreamingSnake, gei.Name)
	valid = reporter.check(gei.Name.Value != unspecifiedEnumName, "%s reversed EnumItem value %s", gei.Name.Pos, gei.Name.Value) && valid
	meta, metaValid := parseDecoratorMeta(reporter, gei.Decorators, _DecoratorContext{
		allowDesc:       true,
		allowDeprecated: true,
	})
	valid = metaValid && valid

	return &schema.EnumItem{
		Pos:              position(gei.Name.Pos),
		Name:             gei.Name.Value,
		Description:      meta.Description,
		Deprecated:       meta.Deprecated,
		DeprecatedReason: meta.DeprecatedReason,
	}, valid
}
