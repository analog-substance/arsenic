package engine

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/analog-substance/tengo/v2"
)

var ErrSignaled error = errors.New("process signaled to close")

// ExecModuleMap represents the 'exec' import module
func (s *Script) ExecModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"err_signaled": &tengo.Error{
			Value: &tengo.String{
				Value: ErrSignaled.Error(),
			},
		},
		"run_with_sig_handler": &tengo.UserFunction{Name: "run_with_sig_handler", Value: s.tengoRunWithSigHandler},
		"cmd":                  &tengo.UserFunction{Name: "cmd", Value: s.tengoCmd},
	}
}

func (s *Script) tengoRunWithSigHandler(args ...tengo.Object) (tengo.Object, error) {
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

	err := s.runWithSigHandler(cmdName, cmdArgs...)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

func (s *Script) tengoCmd(args ...tengo.Object) (tengo.Object, error) {
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

	cmd := exec.CommandContext(context.Background(), cmdName, cmdArgs...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"set_file_stdin": &tengo.UserFunction{
				Name: "set_file_stdin",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) == 0 {
						return toError(tengo.ErrWrongNumArguments), nil
					}

					file, ok := tengo.ToString(args[0])
					if !ok {
						return toError(tengo.ErrInvalidArgumentType{
							Name:     "file",
							Expected: "string",
							Found:    args[0].TypeName(),
						}), nil
					}

					f, err := os.Open(file)
					if err != nil {
						return toError(err), nil
					}
					cmd.Stdin = f

					return nil, nil
				},
			},
			"run": &tengo.UserFunction{
				Name: "run",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					err := s.runCmdWithSigHandler(cmd)
					if err != nil {
						return toError(err), nil
					}

					return nil, nil
				},
			},
		},
	}, nil
}

func (s *Script) runWithSigHandler(name string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), name, args...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return s.runCmdWithSigHandler(cmd)
}

func (s *Script) runCmdWithSigHandler(cmd *exec.Cmd) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// relay trapped signals to the spawned process
	signaled := false
	go func() {
		for sig := range sigs {
			signaled = true
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
		exiterr, ok := err.(*exec.ExitError)
		if !ok {
			return err
		}

		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if !signaled {
				signaled = status.Signaled()
			}
		}
	}

	if signaled {
		return ErrSignaled
	}

	return nil
}
