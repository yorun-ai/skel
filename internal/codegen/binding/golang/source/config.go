package source

import (
	"go.yorun.ai/skel/schema"
)

const configGoFilename = "config.go"

var configImports = []*Import{
	{Path: "reflect"},
	{Path: "go.yorun.ai/vine/core/conf"},
}

var configGoTemplate = loadGoTemplate("config.go.tpl")

func (g *_Gen) genConfigGo() {
	payload := g.buildConfigGoPayload()
	if len(payload.Data) > 0 {
		g.renderGo(configGoFilename, configGoTemplate, payload)
	}
}

func (g *_Gen) buildConfigGoPayload() *DataGoPayload {
	payload := &DataGoPayload{
		PackageName: g.pkgName,
		Data:        make([]*Data, 0, len(g.view.Configs)),
	}
	for _, dataType := range g.view.Configs {
		castedData := g.types.castData(dataType)
		castedData.SpecName = "_" + castedData.Name + "Spec"
		castedData.SkelName = dataType.SkelName
		castedData.Hash = dataType.Hash
		switch dataType.Lifecycle {
		case schema.ConfigLifecycleEternal:
			castedData.Lifecycle = string(dataType.Lifecycle)
			castedData.RegisterFunc = "LifecycleEternal"
		case schema.ConfigLifecycleInstant:
			castedData.Lifecycle = string(dataType.Lifecycle)
			castedData.RegisterFunc = "LifecycleInstant"
		}
		payload.Data = append(payload.Data, castedData)
	}
	imports := buildDataImports(payload.Data)
	imports = append(imports, configImports...)
	payload.StdImports, payload.ModuleImports = splitImports(imports)
	return payload
}
