package engine

import (
	"errors"

	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/spf13/cobra"
)

type CobraCmd struct {
	tengo.ObjectImpl
	Value     *cobra.Command
	objectMap map[string]tengo.Object
	script    *Script
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
	}

	cobraCmd.objectMap = objectMap
	return cobraCmd
}

func (c *CobraCmd) addCommand(args ...tengo.Object) (tengo.Object, error) {
	var cmds []*cobra.Command
	for _, arg := range args {
		c, ok := arg.(*CobraCmd)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "command arg",
				Expected: "cobra-cmd",
				Found:    arg.TypeName(),
			}
		}
		cmds = append(cmds, c.Value)
	}
	c.Value.AddCommand(cmds...)
	return nil, nil
}

func (c *CobraCmd) setRun(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	fn, ok := args[0].(*tengo.CompiledFunction)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "run",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	c.Value.RunE = func(cmd *cobra.Command, args []string) error {
		return c.runCompiledFunction(fn, args)
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
		c.cobraRootCmdPersistentPreRun(cmd, args)
		return c.runCompiledFunction(fn, args)
	}
	return nil, nil
}

func (c *CobraCmd) runCompiledFunction(fn *tengo.CompiledFunction, args []string) error {
	vm := tengo.NewVM(c.script.compiled.Bytecode(), c.script.compiled.Globals(), -1)
	ch := make(chan error, 1)

	errEmpty := errors.New("")

	go func() {
		obj, err := vm.RunCompiled(fn, c, toStringArray(args))
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
	case <-c.script.ctx.Done():
		vm.Abort()
		err = c.script.ctx.Err()
	case err = <-ch:
	}

	if err != nil && err != errEmpty {
		return err
	}

	return nil
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
	disableGit, _ := cmd.Flags().GetBool("disable-git")
	if disableGit {
		c.script.isGit = false
	}
}
