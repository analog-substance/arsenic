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

type Script struct {
	path     string
	name     string
	ctx      context.Context
	cancel   context.CancelFunc
	script   *tengo.Script
	compiled *tengo.Compiled
	args     []string
	isGit    bool
	signaled bool
}

func NewScript(path string) (*Script, error) {
	ctx, cancel := context.WithCancel(context.Background())
	script := &Script{
		path:   path,
		name:   filepath.Base(path),
		ctx:    ctx,
		cancel: cancel,
		isGit:  util.DirExists(".git"),
	}

	// Might want to change this so it just removes any shebang
	shebangRe := regexp.MustCompile(`#!\s*/usr/bin/env arsenic\s*`)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bytes = shebangRe.ReplaceAll(bytes, []byte{})

	s := tengo.NewScript(bytes)
	s.SetImports(script.NewModuleMap())

	s.Add("check_err", &tengo.UserFunction{
		Name:  "check_err",
		Value: script.checkErr,
	})

	script.script = s

	return script, nil
}

func (s *Script) NewModuleMap() *tengo.ModuleMap {
	moduleMap := stdlib.GetModuleMap(stdlib.AllModuleNames()...)

	moduleMap.AddBuiltinModule("filepath", s.FilePathModule())
	moduleMap.AddBuiltinModule("git", s.GitModule())
	moduleMap.AddBuiltinModule("slice", s.SliceModule())
	moduleMap.AddBuiltinModule("url", s.URLModule())
	moduleMap.AddBuiltinModule("arsenic", s.ArsenicModule())
	moduleMap.AddBuiltinModule("script", s.ScriptModule())
	moduleMap.AddBuiltinModule("exec", s.ExecModule())
	moduleMap.AddBuiltinModule("os2", s.OS2Module())
	moduleMap.AddBuiltinModule("set", s.SetModule())
	moduleMap.AddBuiltinModule("cobra", s.CobraModule())
	moduleMap.AddBuiltinModule("nmap", s.NmapModule())
	moduleMap.AddBuiltinModule("scope", s.ScopeModule())
	moduleMap.AddBuiltinModule("log", logModule)

	return moduleMap
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

func (s *Script) Signal() {
	s.signaled = true
	s.stop(ErrSignaled.Error())
}

func (s *Script) runCompiledFunction(fn *tengo.CompiledFunction, args ...tengo.Object) (tengo.Object, error) {
	vm := tengo.NewVM(s.compiled.Bytecode(), s.compiled.Globals(), -1)
	ch := make(chan tengo.Object, 1)

	go func() {
		obj, err := vm.RunCompiled(fn, args...)
		if err != nil {
			ch <- toError(err)
			return
		}

		ch <- obj
	}()

	var obj tengo.Object
	var err error
	select {
	case <-s.ctx.Done():
		vm.Abort()
		err = s.ctx.Err()
	case obj = <-ch:
	}

	if err != nil {
		return nil, err
	}

	errObj, ok := obj.(*tengo.Error)
	if ok {
		return nil, errors.New(errObj.String())
	}

	return obj, nil
}

func (s *Script) checkErr(args ...tengo.Object) (tengo.Object, error) {
	for _, arg := range args {
		errObj, ok := arg.(*tengo.Error)
		if ok {
			s.tengoStop(errObj)
			break
		}
	}

	return nil, nil
}
