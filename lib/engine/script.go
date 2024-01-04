package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/analog-substance/tengo/v2"
	modexec "github.com/analog-substance/tengomod/exec"
	"github.com/analog-substance/tengomod/interop"
)

// ScriptModule represents the 'script' import module
func (s *Script) ScriptModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"path": &tengo.String{
			Value: s.path,
		},
		"name": &tengo.String{
			Value: s.name,
		},
		"stop": &interop.AdvFunction{
			Name:    "stop",
			NumArgs: interop.MaxArgs(1),
			Args:    []interop.AdvArg{interop.ObjectArg("message")},
			Value:   s.tengoStop,
		},
		"fatal": &interop.AdvFunction{
			Name:    "fatal",
			NumArgs: interop.MaxArgs(1),
			Args:    []interop.AdvArg{interop.ObjectArg("message")},
			Value:   s.tengoFatal,
		},
		"run_script": &interop.AdvFunction{
			Name:    "run_script",
			NumArgs: interop.MinArgs(1),
			Args: []interop.AdvArg{
				interop.StrArg("path"),
				interop.StrSliceArg("args", true),
			},
			Value: s.tengoRunScript,
		},
		"run_script_with_sig_handler": &interop.AdvFunction{
			Name:    "run_script_with_sig_handler",
			NumArgs: interop.MinArgs(1),
			Args: []interop.AdvArg{
				interop.StrArg("path"),
				interop.StrSliceArg("args", true),
			},
			Value: s.tengoRunScriptWithSigHandler,
		},
		"find": &interop.AdvFunction{
			Name:    "find",
			NumArgs: interop.ExactArgs(1),
			Args: []interop.AdvArg{
				interop.StrArg("path"),
			},
			Value: s.tengoFindScript,
		},
		"args": &tengo.UserFunction{
			Name: "args",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				return interop.GoStrSliceToTArray(s.args), nil
			},
		},
	}
}

// tengoRunScript is the tengo function version of runScript.
// Represents 'script.run_script(path string, args ...string) error'
func (s *Script) tengoRunScript(args interop.ArgMap) (tengo.Object, error) {
	path, _ := args.GetString("path")
	scriptArgs, _ := args.GetStringSlice("args")

	err := s.runScript(path, scriptArgs...)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}
	return nil, nil
}

// tengoRunScriptWithSigHandler is the tengo function version of runScriptWithSigHandler.
// Represents 'script.run_script_with_sig_handler(path string, args ...string) error'
func (s *Script) tengoRunScriptWithSigHandler(args interop.ArgMap) (tengo.Object, error) {
	path, _ := args.GetString("path")
	scriptArgs, _ := args.GetStringSlice("args")

	err := s.runScriptWithSigHandler(path, scriptArgs...)
	if err != nil {
		return interop.GoErrToTErr(err), nil
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

	script.caller = s

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

	script.caller = s

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
		return modexec.ErrSignaled
	}

	return nil
}

// tengoFindScript is the tengo function version of findScript.
// Represents 'script.find(script string) string|error'
func (s *Script) tengoFindScript(args interop.ArgMap) (tengo.Object, error) {
	path, _ := args.GetString("path")

	fullPath, err := s.findScript(path)
	if err != nil {
		return interop.GoErrToTErr(err), nil
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
func (s *Script) tengoStop(args interop.ArgMap) (tengo.Object, error) {
	message := ""
	obj, ok := args.GetObject("message")
	if ok {
		message, _ = tengo.ToString(obj)
	}

	s.stop(message)
	return nil, nil
}

// tengoFatal is the tengo function version of fatal.
// Represents 'script.fatal(msg string|error)'
func (s *Script) tengoFatal(args interop.ArgMap) (tengo.Object, error) {
	message := ""
	obj, ok := args.GetObject("message")
	if ok {
		message, _ = tengo.ToString(obj)
	}

	s.stop(message)
	return nil, nil
}

// stop prints the message and stops the current script.
func (s *Script) stop(message string) {
	message = strings.ReplaceAll(message, `\n`, "\n")
	message = strings.ReplaceAll(message, `\t`, "\t")

	if message != "" {
		fmt.Println(message)
	}

	go func() {
		s.cancel()
	}()
	time.Sleep(1 * time.Millisecond)
}

func (s *Script) fatal(message string) {
	message = strings.ReplaceAll(message, `\n`, "\n")
	message = strings.ReplaceAll(message, `\t`, "\t")

	if message != "" {
		fmt.Println(message)
	}

	s.err = errors.New(message)

	go func() {
		s.cancel()
	}()
	time.Sleep(1 * time.Millisecond)
}
