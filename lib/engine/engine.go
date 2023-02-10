package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func NewScript(path string) (*Script, error) {
	ctx, cancel := context.WithCancel(context.Background())
	script := &Script{
		ctx:    ctx,
		cancel: cancel,
		isGit:  util.DirExists(".git"),
	}

	shebangRe := regexp.MustCompile(`#!\s*/usr/bin/env arsenic\s*`)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

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

	s.Add("SCRIPT_PATH", path)
	s.Add("SCRIPT_NAME", filepath.Base(path))

	script.script = s

	return script, nil
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

func (s *Script) runCompiledFunction(fn *tengo.CompiledFunction, args ...tengo.Object) error {
	vm := tengo.NewVM(s.compiled.Bytecode(), s.compiled.Globals(), -1)
	ch := make(chan error, 1)

	errEmpty := errors.New("")

	go func() {
		obj, err := vm.RunCompiled(fn, args...)
		if err != nil {
			ch <- err
			return
		}

		errObj, ok := obj.(*tengo.Error)
		if ok {
			ch <- errors.New(errObj.String())
		} else {
			ch <- errEmpty
		}
	}()

	var err error
	select {
	case <-s.ctx.Done():
		vm.Abort()
		err = s.ctx.Err()
	case err = <-ch:
	}

	if err != nil && err != errEmpty {
		return err
	}

	return nil
}
