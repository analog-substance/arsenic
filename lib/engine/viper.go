package engine

import (
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/spf13/viper"
)

// ViperModule represents the 'viper' import module
func (s *Script) ViperModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"get_string": &tengo.UserFunction{Name: "get_string", Value: stdlib.FuncASRS(viper.GetString)},
		"get_int":    &tengo.UserFunction{Name: "get_int", Value: funcASRI(viper.GetInt)},
		"get_bool":   &tengo.UserFunction{Name: "get_bool", Value: funcASRB(viper.GetBool)},
	}
}
