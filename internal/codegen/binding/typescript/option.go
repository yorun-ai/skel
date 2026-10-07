package typescript

import (
	"go.yorun.ai/skel/internal/codegen"
)

type Option struct {
	ApiFilter   codegen.ApiFilter
	AsModule    bool
	Out         string
	Module      string
	Imports     map[string]string
	ModuleScope string
}
