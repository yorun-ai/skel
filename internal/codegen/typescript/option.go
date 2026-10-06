package typescript

import "go.yorun.ai/skel/internal/codegen/common"

type Option struct {
	ApiFilter   common.ApiFilter
	AsModule    bool
	Out         string
	Module      string
	Imports     map[string]string
	ModuleScope string
}
