package engine

import (
	"context"
	"os"
	"path/filepath"

	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/parser"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod"
	modexec "github.com/analog-substance/tengomod/exec"
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
	moduleMap := tengomod.GetModuleMap(
		tengomod.WithContext(s.ctx),
		tengomod.WithCompiledFunc(func() *tengo.Compiled {
			return s.compiled
		}),
	)

	moduleMap.AddMap(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	moduleMap.AddBuiltinModule("git", s.GitModule())
	moduleMap.AddBuiltinModule("arsenic", s.ArsenicModule())
	moduleMap.AddBuiltinModule("script", s.ScriptModule())
	moduleMap.AddBuiltinModule("cobra", s.CobraModule())
	moduleMap.AddBuiltinModule("scope", s.ScopeModule())
	moduleMap.AddBuiltinModule("wordlist", s.WordlistModule())

	return moduleMap
}

func (s *Script) Run(args []string) error {
	s.args = args

	compiled, err := s.script.Compile()
	if err != nil {
		if compilerErr, ok := err.(*tengo.CompilerError); ok {
			s.updateFileSet(compilerErr.FileSet)
		}
		return err
	}

	s.updateFileSet(compiled.Bytecode().FileSet)

	s.compiled = compiled

	return s.compiled.RunContext(s.ctx)
}

func (s *Script) updateFileSet(fileSet *parser.SourceFileSet) {
	for _, file := range fileSet.Files {
		if file.Name == "(main)" {
			file.Name = s.path
			break
		}
	}
}

func (s *Script) Signal() {
	s.signaled = true
	s.stop(modexec.ErrSignaled.Error())
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
