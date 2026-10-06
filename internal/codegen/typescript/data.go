package typescript

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"go.yorun.ai/skel/internal/codegen/common"
	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/util/nameutil"
	"go.yorun.ai/skel/internal/util/sliceutil"
)

const dataTsFilename = "data.ts"

//go:embed tpl/data.ts.tpl
var dataTsTemplate string

type _DataTsPayload struct {
	TypeImports []*_TypeImport
	Enums       []*_Enum
	Data        []*_Data
}

func (g *_Gen) genDataTs() {
	payload := g.buildDataTsPayload()
	g.renderTs(dataTsFilename, dataTsTemplate, payload)
}

func (g *_Gen) buildDataTsPayload() *_DataTsPayload {
	dataList := g.apiView.Data
	payload := &_DataTsPayload{
		TypeImports: g.types.buildDataExternalImports(dataList),
		Enums:       sliceutil.Map(g.apiView.Enums, castEnum),
		Data:        make([]*_Data, 0, len(dataList)),
	}
	for _, dataType := range dataList {
		castedData := g.types.castData(dataType)
		payload.Data = append(payload.Data, castedData)
	}
	return payload
}

func (r _Types) buildDataExternalImports(dataList []*model.Data) []*_TypeImport {
	types := make([]*model.Type, 0)
	for _, dataType := range dataList {
		for _, member := range dataType.Members {
			types = append(types, member.Type)
		}
	}
	return r.buildExternalTypeImports(types)
}

func (r _Types) buildExternalTypeImports(types []*model.Type) []*_TypeImport {
	imports := make([]*_TypeImport, 0)
	seen := make(map[string]struct{})
	common.VisitTypes(types, func(current *model.Type) {
		current = r.bindings.Type(current)
		if current.ExternalImportPath != "" {
			key := current.ExternalAlias + "\x00" + current.ExternalImportPath
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				imports = append(imports, &_TypeImport{Alias: current.ExternalAlias, Path: current.ExternalImportPath})
			}
		}
	})
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].Path == imports[j].Path {
			return imports[i].Alias < imports[j].Alias
		}
		return imports[i].Path < imports[j].Path
	})
	return imports
}

type _Data struct {
	Name         string
	FullName     string
	CommentLines []string
	Members      []*_DataMember
}

func (r _Types) castData(p *model.Data) *_Data {
	data := &_Data{
		Name:         transDataName(p),
		CommentLines: deprecatedTsDocLines(tsCommentLines(p.Description, ""), p.Deprecated, p.DeprecatedReason),
		Members:      make([]*_DataMember, 0, len(p.Members)),
	}
	for _, member := range p.Members {
		castedMember := r.castDataMember(member)
		data.Members = append(data.Members, castedMember)
	}

	data.FullName = data.Name
	if p.TypeParameters != nil {
		tpNames := sliceutil.Map(p.TypeParameters, func(tp *model.TypeParameter) string {
			return tp.Name
		})
		data.FullName = fmt.Sprintf("%s<%s>", data.Name, strings.Join(tpNames, ", "))
	}

	if len(data.Members) > 0 {
		memberNames := sliceutil.Map(data.Members, func(m *_DataMember) string { return m.Name })
		maxMemberNameLen := nameutil.MaxLength(memberNames)
		sliceutil.ForEach(data.Members, func(m *_DataMember) {
			m.NamePadding = nameutil.PaddingSpaces(maxMemberNameLen - len(m.Name))
		})
	}

	return data
}

func transDataName(p *model.Data) string {
	return nameutil.ToCamel(p.Name)
}

type _DataMember struct {
	Name         string
	NamePadding  string
	CommentLines []string
	Type         *_Type
}

func (r _Types) castDataMember(p *model.DataMember) *_DataMember {
	memberType := r.castType(p.Type)
	return &_DataMember{
		Name:         p.Name,
		CommentLines: deprecatedTsDocLines(tsCommentLines(p.Description, p.Example), p.Deprecated, p.DeprecatedReason),
		Type:         memberType,
	}
}
