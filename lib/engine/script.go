package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/analog-substance/tengo/v2"
)

// ScriptModuleMap represents the 'script' import module
func (s *Script) ScriptModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"stop": &tengo.UserFunction{
			Name:  "stop",
			Value: s.tengoStop,
		},
		"run_script": &tengo.UserFunction{
			Name:  "run_script",
			Value: s.tengoRunScript,
		},
		"run_script_with_sig_handler": &tengo.UserFunction{
			Name:  "run_script_with_sig_handler",
			Value: s.tengoRunScriptWithSigHandler,
		},
		"find": &tengo.UserFunction{
			Name:  "find",
			Value: s.tengoFindScript,
		},
		"args": &tengo.UserFunction{
			Name: "args",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				return sliceToStringArray(s.args), nil
			},
		},
	}
}

// tengoRunScript is the tengo function version of runScript.
// Represents 'script.run_script(path string, args ...string) error'
func (s *Script) tengoRunScript(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	scriptArgs := sliceToStringSlice(args[1:])

	err := s.runScript(path, scriptArgs...)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

// tengoRunScriptWithSigHandler is the tengo function version of runScriptWithSigHandler.
// Represents 'script.run_script_with_sig_handler(path string, args ...string) error'
func (s *Script) tengoRunScriptWithSigHandler(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	scriptArgs := sliceToStringSlice(args[1:])

	err := s.runScriptWithSigHandler(path, scriptArgs...)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

// runScript runs the scripts with the args
func (s *Script) runScript(path string, args ...string) error {
	scriptPath, err := s.findScript(path)
	if err != nil {
		return err
	}

	script, err := NewScript(scriptPath)
	if err != nil {
		return err
	}
	return script.Run(args)
}

// runScriptWithSigHandler is like runScript but traps signals like os.Interrupt and syscall.SIGTERM
// and stops the script instead of killing the entire process
func (s *Script) runScriptWithSigHandler(path string, args ...string) error {
	scriptPath, err := s.findScript(path)
	if err != nil {
		return err
	}

	script, err := NewScript(scriptPath)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// relay trapped signals to the spawned process
	signaled := false
	go func() {
		for range sigs {
			signaled = true
			script.Signal()
		}
	}()

	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

	err = script.Run(args)
	if !signaled && (script.signaled || err == context.Canceled) {
		signaled = true
	}

	if signaled {
		return ErrSignaled
	}

	return nil
}

// tengoFindScript is the tengo function version of findScript.
// Represents 'script.find(script string) string|error'
func (s *Script) tengoFindScript(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	fullPath, err := s.findScript(path)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{Value: fullPath}, nil
}

// findScript attempts to get the full path of the specified script
func (s *Script) findScript(path string) (string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}

	fullPath, err := exec.LookPath(path)
	if err == nil {
		return fullPath, nil
	}

	// Look within the script install directory
	return "", os.ErrNotExist
}

// tengoStop is the tengo function version of stop.
// Represents 'script.stop(msg string|error)'
func (s *Script) tengoStop(args ...tengo.Object) (tengo.Object, error) {
	message := ""
	if len(args) == 1 {
		message, _ = tengo.ToString(args[0])
	}

	s.stop(message)
	return nil, nil
}

// stop prints the message and stops the current script.
func (s *Script) stop(args ...string) {
	if len(args) == 1 {
		message := args[0]
		if message != "" {
			message = strings.ReplaceAll(message, `\n`, "\n")
			message = strings.ReplaceAll(message, `\t`, "\t")
			fmt.Println(message)
		}
	}

	go func() {
		s.cancel()
	}()
	time.Sleep(1 * time.Millisecond)
}
