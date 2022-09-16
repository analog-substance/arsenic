package engine

import (
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
			},
		},
		"add_command": &tengo.UserFunction{
			Name:  "add_command",
			Value: cobraCmd.addCommand,
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

	c.Value.Run = func(cmd *cobra.Command, args []string) {
		vm := tengo.NewVM(c.script.compiled.Bytecode(), c.script.compiled.Globals(), -1)
		vm.RunCompiled(fn, c, toStringArray(args))
	}
	return nil, nil
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
	err := c.Value.Execute()
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
