package engine

import (
	"sort"

	"github.com/d5/tengo/v2"
)

var sortModule *SortModule = &SortModule{}

type SortModule struct {
	moduleMap map[string]tengo.Object
}

func (m *SortModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"strings": &tengo.UserFunction{Name: "strings", Value: m.strings},
		}
	}
	return m.moduleMap
}

func (m *SortModule) strings(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	array, ok := args[0].(*tengo.Array)
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "slice",
			Expected: "array",
			Found:    args[0].TypeName(),
		}), nil
	}

	slice, err := toStringSlice(array)
	if err != nil {
		return toError(err), nil
	}

	sort.Strings(slice)

	return toStringArray(slice), nil
}
