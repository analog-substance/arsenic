package script

import (
	"sort"

	"github.com/d5/tengo/v2"
)

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
		return nil, tengo.ErrWrongNumArguments
	}

	array, ok := args[0].(*tengo.Array)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "slice",
			Expected: "array",
			Found:    args[0].TypeName(),
		}
	}

	slice, err := toStringSlice(array)
	if err != nil {
		return nil, err
	}

	sort.Strings(slice)

	return toStringArray(slice), nil
}

var sortModule *SortModule = &SortModule{}
