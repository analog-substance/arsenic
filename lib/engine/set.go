package engine

import (
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

func (s *Script) SetModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"new_string_set": &tengo.UserFunction{Name: "new_string_set", Value: s.newStringSet},
	}
}

func (s *Script) newStringSet(args ...tengo.Object) (tengo.Object, error) {
	stringSet := set.NewStringSet()
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"add": &tengo.UserFunction{
				Name: "add",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					item, ok := tengo.ToString(args[0])
					if !ok {
						return toError(tengo.ErrInvalidArgumentType{
							Name:     "item",
							Expected: "string",
							Found:    args[0].TypeName(),
						}), nil
					}

					value := tengo.FalseValue
					if stringSet.Add(item) {
						value = tengo.TrueValue
					}

					return value, nil
				},
			},
			"sorted_string_slice": &tengo.UserFunction{
				Name:  "sorted_string_slice",
				Value: stdlib.FuncARSs(stringSet.SortedStringSlice),
			},
		},
	}, nil
}
