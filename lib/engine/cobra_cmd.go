package engine

import (
	"strings"

	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/spf13/cobra"
)

type CobraCmd struct {
	tengo.ObjectImpl
	Value                 *cobra.Command
	objectMap             map[string]tengo.Object
	script                *Script
	disableGitFlagEnabled bool
}

func makeCobraCmd(cmd *cobra.Command, script *Script) *CobraCmd {
	cobraCmd := &CobraCmd{
		Value:  cmd,
		script: script,
	}

	objectMap := map[string]tengo.Object{
		"flags": &tengo.ImmutableMap{
			Value: map[string]tengo.Object{
				"boolp": &tengo.UserFunction{
					Name:  "boolp",
					Value: funcASSBSRBp(cobraCmd.Value.Flags().BoolP),
				},
				"bool": &tengo.UserFunction{
					Name:  "bool",
					Value: funcASBSRBp(cobraCmd.Value.Flags().Bool),
				},
				"get_bool": &tengo.UserFunction{
					Name:  "get_bool",
					Value: funcASRBE(cobraCmd.Value.Flags().GetBool),
				},
				"intp": &tengo.UserFunction{
					Name:  "intp",
					Value: funcASSISRIp(cobraCmd.Value.Flags().IntP),
				},
				"int": &tengo.UserFunction{
					Name:  "int",
					Value: funcASISRIp(cobraCmd.Value.Flags().Int),
				},
				"get_int": &tengo.UserFunction{
					Name:  "get_int",
					Value: stdlib.FuncASRIE(cobraCmd.Value.Flags().GetInt),
				},
				"stringp": &tengo.UserFunction{
					Name:  "stringp",
					Value: funcASSSSRSp(cobraCmd.Value.Flags().StringP),
				},
				"string": &tengo.UserFunction{
					Name:  "string",
					Value: funcASSSRSp(cobraCmd.Value.Flags().String),
				},
				"get_string": &tengo.UserFunction{
					Name:  "get_string",
					Value: stdlib.FuncASRSE(cobraCmd.Value.Flags().GetString),
				},
				"string_slicep": &tengo.UserFunction{
					Name:  "string_slicep",
					Value: funcASSSsSRSsp(cobraCmd.Value.Flags().StringSliceP),
				},
				"string_slice": &tengo.UserFunction{
					Name:  "string_slice",
					Value: funcASSsSRSsp(cobraCmd.Value.Flags().StringSlice),
				},
				"get_string_slice": &tengo.UserFunction{
					Name:  "get_string_slice",
					Value: funcASRSsE(cobraCmd.Value.Flags().GetStringSlice),
				},
			},
		},
		"persistent_flags": &tengo.ImmutableMap{
			Value: map[string]tengo.Object{
				"boolp": &tengo.UserFunction{
					Name:  "boolp",
					Value: funcASSBSRBp(cobraCmd.Value.PersistentFlags().BoolP),
				},
				"bool": &tengo.UserFunction{
					Name:  "bool",
					Value: funcASBSRBp(cobraCmd.Value.PersistentFlags().Bool),
				},
				"intp": &tengo.UserFunction{
					Name:  "intp",
					Value: funcASSISRIp(cobraCmd.Value.PersistentFlags().IntP),
				},
				"int": &tengo.UserFunction{
					Name:  "int",
					Value: funcASISRIp(cobraCmd.Value.PersistentFlags().Int),
				},
				"stringp": &tengo.UserFunction{
					Name:  "stringp",
					Value: funcASSSSRSp(cobraCmd.Value.PersistentFlags().StringP),
				},
				"string": &tengo.UserFunction{
					Name:  "string",
					Value: funcASSSRSp(cobraCmd.Value.PersistentFlags().String),
				},
				"string_slicep": &tengo.UserFunction{
					Name:  "string_slicep",
					Value: funcASSSsSRSsp(cobraCmd.Value.PersistentFlags().StringSliceP),
				},
				"string_slice": &tengo.UserFunction{
					Name:  "string_slice",
					Value: funcASSsSRSsp(cobraCmd.Value.PersistentFlags().StringSlice),
				},
			},
		},
		"add_command": &tengo.UserFunction{
			Name:  "add_command",
			Value: cobraCmd.addCommand,
		},
		"set_persistent_pre_run": &tengo.UserFunction{
			Name:  "set_persistent_pre_run",
			Value: cobraCmd.setPersistentPreRun,
		},
		"set_run": &tengo.UserFunction{
			Name:  "set_run",
			Value: cobraCmd.setRun,
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
		"register_flag_completion_func": &tengo.UserFunction{
			Name:  "register_flag_completion_func",
			Value: cobraCmd.registerFlagCompletionFunc,
		},
	}

	cobraCmd.objectMap = objectMap
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

func (c *CobraCmd) setRun(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	fn, ok := args[0].(*tengo.CompiledFunction)
	if ok {
		c.Value.RunE = func(cmd *cobra.Command, args []string) error {
			_, err := c.script.runCompiledFunction(fn, makeCobraCmd(cmd, c.script), sliceToStringArray(args))
			return err
		}
	} else {
		defaultCmd, ok := args[0].(*CobraCmd)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "run",
				Expected: "func|cobra-cmd",
				Found:    args[0].TypeName(),
			}
		}

		c.Value.RunE = func(cmd *cobra.Command, args []string) error {
			commandPathArgs := strings.Split(defaultCmd.Value.CommandPath(), " ")[1:]
			cmd.Root().SetArgs(append(commandPathArgs, args...))
			return defaultCmd.Value.Execute()
		}
	}

	return nil, nil
}

func (c *CobraCmd) setPersistentPreRun(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	fn, ok := args[0].(*tengo.CompiledFunction)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "persistent_pre_run",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	c.Value.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// If completion command isn't disabled and the current command is one of the "completion" sub commands
		// or __complete, don't run the set persistent pre run
		if !cmd.CompletionOptions.DisableDefaultCmd &&
			(cmd.Name() == "__complete" || (cmd.HasParent() && cmd.Parent().Name() == "completion")) {
			return nil
		}

		c.cobraRootCmdPersistentPreRun(cmd, args)
		_, err := c.script.runCompiledFunction(fn, makeCobraCmd(cmd, c.script), sliceToStringArray(args))
		return err
	}
	return nil, nil
}

func (c *CobraCmd) registerFlagCompletionFunc(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	flag, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "flag",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	fn, ok := args[1].(*tengo.CompiledFunction)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "persistent_pre_run",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	c.Value.RegisterFlagCompletionFunc(flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		res, err := c.script.runCompiledFunction(fn, makeCobraCmd(cmd, c.script), sliceToStringArray(args), &tengo.String{Value: toComplete})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		_, ok := res.(*tengo.Error)
		if ok {
			return nil, cobra.ShellCompDirectiveError
		}

		slice, err := arrayToStringSlice(res.(*tengo.Array))
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
		return toError(err), nil
	}
	return nil, nil
}

// CanCall should return whether the Object can be Called.
func (c *CobraCmd) CanCall() bool {
	return true
}

func (c *CobraCmd) IndexGet(index tengo.Object) (tengo.Object, error) {
	strIdx, ok := tengo.ToString(index)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}

	res, ok := c.objectMap[strIdx]
	if !ok {
		res = tengo.UndefinedValue
	}
	return res, nil
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
