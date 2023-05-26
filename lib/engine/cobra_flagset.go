package engine

import (
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod/interop"
	"github.com/analog-substance/tengomod/types"
	"github.com/spf13/pflag"
)

type CobraFlagSet struct {
	types.PropObject
	Value  *pflag.FlagSet
	script *Script
}

func makeCobraFlagSet(flagSet *pflag.FlagSet, script *Script) *CobraFlagSet {
	cobraFlagSet := &CobraFlagSet{
		Value:  flagSet,
		script: script,
	}

	objectMap := map[string]tengo.Object{
		"boolp": &tengo.UserFunction{
			Name:  "boolp",
			Value: interop.FuncASSBSRBp(flagSet.BoolP),
		},
		"bool": &tengo.UserFunction{
			Name:  "bool",
			Value: interop.FuncASBSRBp(flagSet.Bool),
		},
		"get_bool": &tengo.UserFunction{
			Name:  "get_bool",
			Value: interop.FuncASRBE(flagSet.GetBool),
		},
		"intp": &tengo.UserFunction{
			Name:  "intp",
			Value: interop.FuncASSISRIp(flagSet.IntP),
		},
		"int": &tengo.UserFunction{
			Name:  "int",
			Value: interop.FuncASISRIp(flagSet.Int),
		},
		"get_int": &tengo.UserFunction{
			Name:  "get_int",
			Value: stdlib.FuncASRIE(flagSet.GetInt),
		},
		"stringp": &tengo.UserFunction{
			Name:  "stringp",
			Value: interop.FuncASSSSRSp(flagSet.StringP),
		},
		"string": &tengo.UserFunction{
			Name:  "string",
			Value: interop.FuncASSSRSp(flagSet.String),
		},
		"get_string": &tengo.UserFunction{
			Name:  "get_string",
			Value: stdlib.FuncASRSE(flagSet.GetString),
		},
		"string_slicep": &tengo.UserFunction{
			Name:  "string_slicep",
			Value: interop.FuncASSSsSRSsp(flagSet.StringSliceP),
		},
		"string_slice": &tengo.UserFunction{
			Name:  "string_slice",
			Value: interop.FuncASSsSRSsp(flagSet.StringSlice),
		},
		"get_string_slice": &tengo.UserFunction{
			Name:  "get_string_slice",
			Value: interop.FuncASRSsE(flagSet.GetStringSlice),
		},
	}

	properties := map[string]types.Property{
		"sort_flags": {
			Get: func() tengo.Object {
				return interop.GoBoolToTBool(flagSet.SortFlags)
			},
			Set: func(o tengo.Object) error {
				b1, err := interop.TBoolToGoBool(o, "sort_flags")
				if err != nil {
					return err
				}

				flagSet.SortFlags = b1
				return nil
			},
		},
	}

	cobraFlagSet.PropObject = types.PropObject{
		ObjectMap:  objectMap,
		Properties: properties,
	}

	return cobraFlagSet
}

// TypeName should return the name of the type.
func (c *CobraFlagSet) TypeName() string {
	return "cobra-flagset"
}

// String should return a string representation of the type's value.
func (c *CobraFlagSet) String() string {
	return "<cobra-flagset>"
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (c *CobraFlagSet) IsFalsy() bool {
	return c.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (c *CobraFlagSet) CanIterate() bool {
	return false
}
