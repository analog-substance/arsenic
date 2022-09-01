package script

import (
	"context"
	"os"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
)

var moduleMap *tengo.ModuleMap

func Run(path string) error {
	bytes, _ := os.ReadFile(path)
	script := tengo.NewScript(bytes)

	ctx, cancel := context.WithCancel(context.Background())
	engineModule.stopScript = cancel

	script.SetImports(moduleMap)
	_, err := script.RunContext(ctx)
	return err
}

func init() {
	moduleMap = stdlib.GetModuleMap(stdlib.AllModuleNames()...)

	moduleMap.AddBuiltinModule("filepath", filepathModule.ModuleMap())
	moduleMap.AddBuiltinModule("git", gitModule.ModuleMap())
	moduleMap.AddBuiltinModule("sort", sortModule.ModuleMap())
	moduleMap.AddBuiltinModule("url", urlModule.ModuleMap())
	moduleMap.AddBuiltinModule("arsenic", arsenicModule.ModuleMap())
	moduleMap.AddBuiltinModule("engine", engineModule.ModuleMap())
	moduleMap.AddBuiltinModule("log", logModule)
}
