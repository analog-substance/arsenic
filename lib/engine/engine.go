package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod"
	modexec "github.com/analog-substance/tengomod/exec"
	"github.com/analog-substance/tengomod/interop"
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
		isGit:  fileutil.DirExists(".git"),
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(bytes) > 1 && string(bytes[:2]) == "#!" {
		copy(bytes, "//")
	}

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
	moduleMap := tengomod.GetModuleMap(tengomod.WithCompiledFunc(func() (*tengo.Compiled, context.Context) {
		return s.compiled, s.ctx
	}))

	moduleMap.AddMap(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	moduleMap.AddBuiltinModule("git", s.GitModule())
	moduleMap.AddBuiltinModule("arsenic", s.ArsenicModule())
	moduleMap.AddBuiltinModule("script", s.ScriptModule())
	moduleMap.AddBuiltinModule("cobra", s.CobraModule())
	moduleMap.AddBuiltinModule("scope", s.ScopeModule())
	moduleMap.AddBuiltinModule("log", s.LogModule())
	moduleMap.AddBuiltinModule("ffuf", s.FfufModule())
	moduleMap.AddBuiltinModule("wordlist", s.WordlistModule())

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
	s.stop(modexec.ErrSignaled.Error())
}

func (s *Script) runCompiledFunction(fn *tengo.CompiledFunction, args ...tengo.Object) (tengo.Object, error) {
	vm := tengo.NewVM(s.compiled.Bytecode(), s.compiled.Globals(), -1)
	ch := make(chan tengo.Object, 1)

	go func() {
		obj, err := vm.RunCompiled(fn, args...)
		if err != nil {
			ch <- interop.GoErrToTErr(err)
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
