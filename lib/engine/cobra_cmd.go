package engine

import (
	"strings"

	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod/interop"
	"github.com/analog-substance/tengomod/types"
	"github.com/spf13/cobra"
)

type CobraCmd struct {
	types.PropObject
	Value                 *cobra.Command
	script                *Script
	disableGitFlagEnabled bool
}

func makeCobraCmd(cmd *cobra.Command, script *Script) *CobraCmd {
	cobraCmd := &CobraCmd{
		Value:  cmd,
		script: script,
	}

	objectMap := map[string]tengo.Object{
		"add_command": &tengo.UserFunction{
			Name:  "add_command",
			Value: cobraCmd.addCommand,
		},
		"new_command": &interop.AdvFunction{
			Name:    "new_command",
			NumArgs: interop.ArgRange(1, 2),
			Args:    []interop.AdvArg{interop.StrArg("use"), interop.StrArg("short-description")},
			Value:   cobraCmd.newCommand,
		},
		"set_persistent_pre_run": &interop.AdvFunction{
			Name:    "set_persistent_pre_run",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.CompileFuncArg("persistent-pre-run")},
			Value:   cobraCmd.setPersistentPreRun,
		},
		"set_run": &interop.AdvFunction{
			Name:    "set_run",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.UnionArg("run", interop.CompileFuncType, interop.CustomType(&CobraCmd{}))},
			Value:   cobraCmd.setRun,
		},
		"name": &tengo.UserFunction{
			Name:  "name",
			Value: stdlib.FuncARS(cobraCmd.Value.Name),
		},
		"has_parent": &tengo.UserFunction{
			Name:  "has_parent",
			Value: stdlib.FuncARB(cobraCmd.Value.HasParent),
		},
		"parent": &tengo.UserFunction{
			Name: "parent",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				return makeCobraCmd(cobraCmd.Value.Parent(), cobraCmd.script), nil
			},
		},
		"called_as": &tengo.UserFunction{
			Name:  "called_as",
			Value: stdlib.FuncARS(cobraCmd.Value.CalledAs),
		},
		"enable_completion": &tengo.UserFunction{
			Name: "enable_completion",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if !cobraCmd.isRootCmd() {
					return nil, nil
				}

				cobraCmd.Value.CompletionOptions.DisableDefaultCmd = false
				return nil, nil
			},
		},
		"add_disable_git_flag": &tengo.UserFunction{
			Name: "add_disable_git_flag",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if !cobraCmd.isRootCmd() {
					return nil, nil
				}

				cobraCmd.disableGitFlagEnabled = true
				cobraCmd.Value.PersistentFlags().Bool("disable-git", false, "Disable git commands through the git module.")
				return nil, nil
			},
		},
		"register_flag_completion_func": &interop.AdvFunction{
			Name:    "register_flag_completion_func",
			NumArgs: interop.ExactArgs(2),
			Args:    []interop.AdvArg{interop.StrArg("flag"), interop.CompileFuncArg("fn")},
			Value:   cobraCmd.registerFlagCompletionFunc,
		},
	}
	properties := map[string]types.Property{
		"flags": {
			Get: func() tengo.Object {
				return makeCobraFlagSet(cobraCmd.Value.Flags(), cobraCmd.script)
			},
		},
		"persistent_flags": {
			Get: func() tengo.Object {
				return makeCobraFlagSet(cobraCmd.Value.PersistentFlags(), cobraCmd.script)
			},
		},
	}

	cobraCmd.PropObject = types.PropObject{
		ObjectMap:  objectMap,
		Properties: properties,
	}

	return cobraCmd
}

func (c *CobraCmd) addCommand(args ...tengo.Object) (tengo.Object, error) {
	var cmds []*cobra.Command
	for _, arg := range args {
		cmd, ok := arg.(*CobraCmd)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "command arg",
				Expected: "cobra-cmd",
				Found:    arg.TypeName(),
			}
		}
		cmds = append(cmds, cmd.Value)
	}
	c.Value.AddCommand(cmds...)
	return nil, nil
}

func (c *CobraCmd) newCommand(args interop.ArgMap) (tengo.Object, error) {
	cobraCmd, err := c.script.cobraCmd(args)
	if err != nil {
		return nil, err
	}
	return c.addCommand(cobraCmd)
}

func (c *CobraCmd) setRun(args interop.ArgMap) (tengo.Object, error) {
	if fn, ok := args.GetCompiledFunc("run"); ok {
		c.Value.RunE = func(cmd *cobra.Command, args []string) error {
			runner := interop.NewCompiledFuncRunner(fn, c.script.compiled, c.script.ctx)
			_, err := runner.Run(makeCobraCmd(cmd, c.script), interop.GoStrSliceToTArray(args))
			return err
		}
	} else {
		defaultCmd := args["run"].(*CobraCmd)
		c.Value.RunE = func(cmd *cobra.Command, args []string) error {
			commandPathArgs := strings.Split(defaultCmd.Value.CommandPath(), " ")[1:]
			cmd.Root().SetArgs(append(commandPathArgs, args...))
			return defaultCmd.Value.Execute()
		}
	}

	return nil, nil
}

func (c *CobraCmd) setPersistentPreRun(args interop.ArgMap) (tengo.Object, error) {
	fn, _ := args.GetCompiledFunc("persistent-pre-run")

	c.Value.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If completion command isn't disabled and the current command is one of the "completion" sub commands
		// or __complete, don't run the set persistent pre run
		if !cmd.CompletionOptions.DisableDefaultCmd &&
			(cmd.Name() == "__complete" || (cmd.HasParent() && cmd.Parent().Name() == "completion")) {
			return nil
		}

		c.cobraRootCmdPersistentPreRun(cmd, args)
		runner := interop.NewCompiledFuncRunner(fn, c.script.compiled, c.script.ctx)
		_, err := runner.Run(makeCobraCmd(cmd, c.script), interop.GoStrSliceToTArray(args))
		return err
	}
	return nil, nil
}

func (c *CobraCmd) registerFlagCompletionFunc(args interop.ArgMap) (tengo.Object, error) {
	flag, _ := args.GetString("flag")
	fn, _ := args.GetCompiledFunc("fn")

	c.Value.RegisterFlagCompletionFunc(flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		runner := interop.NewCompiledFuncRunner(fn, c.script.compiled, c.script.ctx)
		res, err := runner.Run(makeCobraCmd(cmd, c.script), interop.GoStrSliceToTArray(args), &tengo.String{Value: toComplete})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		_, ok := res.(*tengo.Error)
		if ok {
			return nil, cobra.ShellCompDirectiveError
		}

		slice, err := interop.TArrayToGoStrSlice(res, "")
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return slice, cobra.ShellCompDirectiveDefault
	})

	return nil, nil
}

func (c *CobraCmd) isRootCmd() bool {
	return c.Value.Annotations["type"] == "root"
}

// TypeName should return the name of the type.
func (c *CobraCmd) TypeName() string {
	return "cobra-cmd"
}

// String should return a string representation of the type's value.
func (c *CobraCmd) String() string {
	return c.Value.Name()
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (c *CobraCmd) IsFalsy() bool {
	return c.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (c *CobraCmd) CanIterate() bool {
	return false
}

// Call should take an arbitrary number of arguments and returns a return
// value and/or an error, which the VM will consider as a run-time error.
func (c *CobraCmd) Call(args ...tengo.Object) (tengo.Object, error) {
	err := c.Value.ExecuteContext(c.script.ctx)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}
	return nil, nil
}

// CanCall should return whether the Object can be Called.
func (c *CobraCmd) CanCall() bool {
	return true
}

func (c *CobraCmd) cobraRootCmdPersistentPreRun(cmd *cobra.Command, args []string) {
	if !c.isRootCmd() {
		return
	}

	if c.disableGitFlagEnabled {
		disableGit, _ := cmd.Flags().GetBool("disable-git")
		if disableGit {
			c.script.isGit = false
		}
	}
}
