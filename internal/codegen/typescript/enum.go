package typescript

import (
	"fmt"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
)

type _Enum struct {
	Name         string
	CommentLines []string
	Items        []*_EnumItem
}

func castEnum(p *model.Enum) *_Enum {
	enum := &_Enum{
		Name:         nameutil.ToCamel(p.Name),
		CommentLines: deprecatedTsDocLines(tsCommentLines(p.Description, ""), p.Deprecated, p.DeprecatedReason),
		Items:        make([]*_EnumItem, 0, len(p.Items)+1),
	}
	enum.Items = append(enum.Items, castEnumItem(p.UnspecifiedItem))
	enum.Items = append(enum.Items, sliceutil.Map(p.Items, castEnumItem)...)
	if len(enum.Items) > 0 {
		itemValues := sliceutil.Map(enum.Items, func(i *_EnumItem) string { return i.Literal })
		maxValueLen := nameutil.MaxLength(itemValues)
		sliceutil.ForEach(enum.Items, func(i *_EnumItem) {
			i.ValuePadding = nameutil.PaddingSpaces(maxValueLen - len(i.Literal))
		})
	}

	return enum
}

func transEnumName(p *model.Enum) string {
	return nameutil.ToCamel(p.Name)
}

type _EnumItem struct {
	Literal      string
	Value        string
	ValuePadding string
	CommentLines []string
}

func castEnumItem(p *model.EnumItem) *_EnumItem {
	return &_EnumItem{
		Literal:      fmt.Sprintf(`"%s"`, nameutil.ToScreamingSnake(p.Name)),
		Value:        nameutil.ToScreamingSnake(p.Name),
		CommentLines: deprecatedTsDocLines(tsCommentLines(common.MergeDescriptionAndExample(p.Description, ""), ""), p.Deprecated, p.DeprecatedReason),
	}
}
