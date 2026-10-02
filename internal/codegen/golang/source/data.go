package source

import (
	"fmt"
	"strings"

	"go.yorun.ai/skelc/internal/codegen/common"
	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/skelmeta"
	"go.yorun.ai/skelc/internal/util/nameutil"
	"go.yorun.ai/skelc/internal/util/sliceutil"
)

const dataGoFilename = "data.go"

var dataGoTemplate = joinTemplates("imports.go.tpl", "data.go.tpl")

type DataGoPayload struct {
	PackageName   string
	StdImports    []*Import
	ModuleImports []*Import
	Data          []*Data
}

func (g *_Gen) genDataGo() {
	payload := g.buildDataGoPayload()
	if len(payload.Data) > 0 {
		g.renderGo(dataGoFilename, dataGoTemplate, payload)
	}
}

func (g *_Gen) buildDataGoPayload() *DataGoPayload {
	payload := &DataGoPayload{
		PackageName: g.pkgName,
		Data:        make([]*Data, 0, len(g.view.Data)),
	}
	for _, dataType := range g.view.Data {
		castedData := g.types.castData(dataType)
		payload.Data = append(payload.Data, castedData)
	}
	imports := g.apiImports(buildDataImports(payload.Data))
	payload.StdImports, payload.ModuleImports = splitImports(imports)

	return payload
}

type Data struct {
	Name             string
	FullName         string
	ReceiverType     string
	ImplName         string
	ConstructorName  string
	SpecName         string
	SkelName         string
	Hash             string
	Lifecycle        string
	RegisterFunc     string
	CommentLines     []string
	Members          []*DataMember
	Sensitive        bool
	MarkerMethodName string
}

func (r _Types) castData(p *model.Data) *Data {
	data := &Data{
		Name:             transDataName(p),
		ImplName:         "_" + transDataName(p),
		ConstructorName:  "_New" + transDataName(p),
		CommentLines:     deprecatedGoDocLines(goDocLines(transDataName(p), p.Description), transDataName(p), p.DeprecatedReason),
		Members:          make([]*DataMember, 0, len(p.Members)),
		Sensitive:        p.Sensitive,
		MarkerMethodName: skelmeta.SensitiveMarkerMethodName,
	}
	for _, member := range p.Members {
		castedMember := r.castDataMember(member)
		data.Members = append(data.Members, castedMember)
	}

	data.FullName = data.Name
	data.ReceiverType = data.Name
	if p.TypeParameters != nil {
		tpNames := sliceutil.Map(p.TypeParameters, func(tp *model.TypeParameter) string {
			return tp.Name
		})
		data.FullName = fmt.Sprintf("%s[%s any]", data.Name, strings.Join(tpNames, ", "))
		data.ReceiverType = fmt.Sprintf("%s[%s]", data.Name, strings.Join(tpNames, ", "))
	}
	return data
}

func transDataName(p *model.Data) string {
	return nameutil.ToCamel(p.Name)
}

type DataMember struct {
	Name         string
	CommentLines []string
	Type         *Type
	SkelName     string
	Sensitive    bool
	Identifier   bool
}

func (r _Types) castDataMember(p *model.DataMember) *DataMember {
	memberType := r.castType(p.Type)
	return &DataMember{
		Name: nameutil.ToCamel(p.Name),
		CommentLines: deprecatedGoDocLines(
			goDocLines(nameutil.ToCamel(p.Name), common.MergeDescriptionAndExample(p.Description, p.Example)),
			nameutil.ToCamel(p.Name),
			p.DeprecatedReason,
		),
		Type:      memberType,
		SkelName:  p.Name,
		Sensitive: p.Sensitive,
	}
}

func buildDataImports(dataList []*Data) []*Import {
	imports := newImportSet()
	for _, data := range dataList {
		for _, member := range data.Members {
			imports.addMany(collectTypeImports(member.Type))
		}
	}
	return imports.sortedValues()
}
