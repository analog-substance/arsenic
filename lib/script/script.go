package script

import (
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
)

var moduleMap *tengo.ModuleMap

func init() {
	moduleMap = stdlib.GetModuleMap(stdlib.AllModuleNames()...)

	moduleMap.AddBuiltinModule("filepath", filepathModule)
	moduleMap.AddBuiltinModule("git", gitModule)
	moduleMap.AddBuiltinModule("log", logModule)
}
