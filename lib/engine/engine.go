package engine

import (
	"context"
	"os"
	"regexp"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

const checkErrSrcModule = `
script := import("script")

export func(err) {
	if is_error(err) {
		script.stop(err)
	}
}
`

type Script struct {
	ctx      context.Context
	cancel   context.CancelFunc
	script   *tengo.Script
	compiled *tengo.Compiled
	args     []string
	isGit    bool
}

func NewScript(path string) *Script {
	ctx, cancel := context.WithCancel(context.Background())
	script := &Script{
		ctx:    ctx,
		cancel: cancel,
		isGit:  util.DirExists(".git"),
	}

	shebangRe := regexp.MustCompile(`#!\s*/usr/bin/env arsenic\s*`)
	bytes, _ := os.ReadFile(path)
	bytes = shebangRe.ReplaceAll(bytes, []byte{})

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
	moduleMap.AddBuiltinModule("set", script.SetModuleMap())
	moduleMap.AddBuiltinModule("cobra", script.CobraModuleMap())
	moduleMap.AddBuiltinModule("nmap", script.NmapModuleMap())
	moduleMap.AddBuiltinModule("log", logModule)
	moduleMap.AddSourceModule("check_err", []byte(checkErrSrcModule))

	s.SetImports(moduleMap)
	script.script = s

	return script
}

func (s *Script) Run(args []string) error {
	s.args = args

	compiled, err := s.script.Compile()
	if err != nil {
		return err
	}

	s.compiled = compiled

	return s.compiled.RunContext(s.ctx)
}
