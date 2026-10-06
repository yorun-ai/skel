package source

import (
	"fmt"

	"go.yorun.ai/skel/internal/util/nameutil"
)

func er(name string) string {
	return fmt.Sprintf("%sER", name)
}

type _ServiceNames struct {
	Name                    string
	SpecName                string
	BaseName                string
	ServerName              string
	DefaultServerName       string
	ClientName              string
	ClientImplName          string
	ClientCtorName          string
	ERBaseName              string
	ERServerName            string
	WrapperERServerName     string
	WrapperERServerCtorName string
	DefaultERServerName     string
	ERClientName            string
	ERClientImplName        string
	ERClientCtorName        string
}

func buildServiceNames(serviceName string) *_ServiceNames {
	names := &_ServiceNames{
		Name: nameutil.ToCamel(serviceName),
	}
	names.SpecName = fmt.Sprintf("_%sSpec", names.Name)
	names.BaseName = fmt.Sprintf("_%s", names.Name)
	names.ServerName = fmt.Sprintf("%sServer", names.Name)
	names.DefaultServerName = fmt.Sprintf("Default%s", names.ServerName)
	names.ClientName = fmt.Sprintf("%sClient", names.Name)
	names.ClientImplName = fmt.Sprintf("_%s", names.ClientName)
	names.ClientCtorName = fmt.Sprintf("New%s", names.ClientName)
	names.ERBaseName = fmt.Sprintf("_%s", er(names.Name))
	names.ERServerName = er(names.ServerName)
	names.WrapperERServerName = fmt.Sprintf("_Wrapper%s", names.ERServerName)
	names.WrapperERServerCtorName = fmt.Sprintf("_NewWrapper%s", names.ERServerName)
	names.DefaultERServerName = er(names.DefaultServerName)
	names.ERClientName = er(names.ClientName)
	names.ERClientImplName = er(names.ClientImplName)
	names.ERClientCtorName = er(names.ClientCtorName)
	return names
}
