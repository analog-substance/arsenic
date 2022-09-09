package engine

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/d5/tengo/v2"
)

var execModule *ExecModule = &ExecModule{}

type ExecModule struct {
	moduleMap map[string]tengo.Object
}

func (m *ExecModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"run_with_sig_handler": &tengo.UserFunction{Name: "run_with_sig_handler", Value: m.tengoRunWithSigHandler},
		}
	}
	return m.moduleMap
}

func (m *ExecModule) tengoRunWithSigHandler(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	cmdName, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "cmd name",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	var cmdArgs []string
	for _, arg := range args[1:] {
		cmdArg, ok := tengo.ToString(arg)
		if !ok {
			return toError(tengo.ErrInvalidArgumentType{
				Name:     "cmd arg",
				Expected: "string",
				Found:    arg.TypeName(),
			}), nil
		}

		cmdArgs = append(cmdArgs, cmdArg)
	}

	err := m.runWithSigHandler(cmdName, cmdArgs...)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

func (m *ExecModule) runWithSigHandler(name string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), name, args...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// relay trapped signals to the spawned process
	go func() {
		for sig := range sigs {
			cmd.Process.Signal(sig)
		}
	}()

	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return err
		}
	}

	return nil
}
