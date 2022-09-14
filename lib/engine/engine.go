package engine

import (
	"context"
	"os"
	"path/filepath"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/spf13/viper"
)

type Script struct {
	ctx    context.Context
	cancel context.CancelFunc
	script *tengo.Script
	isGit  bool
}

func NewScript(path string) *Script {
	if filepath.Ext(path) != ".tengo" {
		path = path + ".tengo"
	}

	ctx, cancel := context.WithCancel(context.Background())
	script := &Script{
		ctx:    ctx,
		cancel: cancel,
		isGit:  util.DirExists(".git"),
	}

	bytes, _ := os.ReadFile(filepath.Join(viper.GetString("scripts-directory"), path))
	s := tengo.NewScript(bytes)

	moduleMap := stdlib.GetModuleMap(stdlib.AllModuleNames()...)

	moduleMap.AddBuiltinModule("filepath", script.FilePathModuleMap())
	moduleMap.AddBuiltinModule("git", script.GitModuleMap())
	moduleMap.AddBuiltinModule("slice", script.SliceModuleMap())
	moduleMap.AddBuiltinModule("url", script.URLModuleMap())
	moduleMap.AddBuiltinModule("arsenic", script.ArsenicModuleMap())
	moduleMap.AddBuiltinModule("script", script.ScriptModuleMap())
	moduleMap.AddBuiltinModule("exec", script.ExecModuleMap())
	moduleMap.AddBuiltinModule("os2", script.OS2ModuleMap())
	moduleMap.AddBuiltinModule("log", logModule)

	s.SetImports(moduleMap)
	script.script = s

	return script
}

func (s *Script) Run(scriptArgs map[string]string) error {
	args := make(map[string]interface{})
	for key, value := range scriptArgs {
		args[key] = value
	}

	err := s.script.Add("args", args)
	if err != nil {
		return err
	}

	_, err = s.script.RunContext(s.ctx)
	return err
}
