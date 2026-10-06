package source

import (
	"strings"

	"go.yorun.ai/skel/internal/codegen/binding"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
)

const enumGoFilename = "enum.go"

var enumGoTemplate = loadGoTemplate("enum.go.tpl")

type EnumGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Enums         []*Enum
}

func (g *_Gen) genEnumGo() {
	payload := g.buildEnumGoPayload()
	if len(payload.Enums) > 0 {
		g.renderGo(enumGoFilename, enumGoTemplate, payload)
	}
}

func (g *_Gen) buildEnumGoPayload() *EnumGoPayload {
	p := &EnumGoPayload{
		PackageName: g.pkgName,
		Enums:       sliceutil.Map(g.view.Enums, castEnum),
	}
	if len(p.Enums) > 0 {
		imports := []*Import{{Path: "fmt"}}
		p.StdImports, p.ModuleImports = splitImports(imports)
	}
	return p
}

type Enum struct {
	Name            string
	VarName         string
	CommentLines    []string
	UnspecifiedItem *EnumItem
	Items           []*EnumItem
}

func castEnum(p *model.Enum) *Enum {
	enum := &Enum{
		Name:            transEnumName(p),
		VarName:         nameutil.ToLowerCamel(p.Name),
		UnspecifiedItem: castEnumItem(p.UnspecifiedItem),
		Items:           sliceutil.Map(p.Items, castEnumItem),
	}
	enum.CommentLines = deprecatedGoDocLines(goDocLines(enum.Name, p.Description), enum.Name, p.DeprecatedReason)

	if len(enum.Items) > 0 {
		enum.UnspecifiedItem.Name = enum.Name + enum.UnspecifiedItem.Name
		enum.UnspecifiedItem.CommentLines = deprecatedGoDocLines(
			goDocLines(enum.UnspecifiedItem.Name, binding.MergeDescriptionAndExample(p.UnspecifiedItem.Description, "")),
			enum.UnspecifiedItem.Name,
			enum.UnspecifiedItem.DeprecatedReason,
		)
		sliceutil.ForEach(enum.Items, func(i *EnumItem) {
			i.Name = enum.Name + i.Name
			i.CommentLines = deprecatedGoDocLines(
				goDocLines(i.Name, binding.MergeDescriptionAndExample(i.Description, "")),
				i.Name,
				i.DeprecatedReason,
			)
		})
	}

	return enum
}

func transEnumName(p *model.Enum) string {
	return nameutil.ToCamel(p.Name)
}

func transUnspecifiedItemName(p *model.Enum) string {
	return transEnumName(p) + castEnumItem(p.UnspecifiedItem).Name
}

type EnumItem struct {
	Name             string
	Value            string
	Description      string
	DeprecatedReason string
	CommentLines     []string
}

func castEnumItem(p *model.EnumItem) *EnumItem {
	name := nameutil.ToCamel(strings.ToLower(p.Name))
	return &EnumItem{
		Name:             name,
		Value:            nameutil.ToScreamingSnake(p.Name),
		Description:      p.Description,
		DeprecatedReason: p.DeprecatedReason,
		CommentLines:     goDocLines(name, binding.MergeDescriptionAndExample(p.Description, "")),
	}
}
