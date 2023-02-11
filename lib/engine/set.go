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
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "item",
							Expected: "string",
							Found:    args[0].TypeName(),
						}
					}

					value := tengo.FalseValue
					if stringSet.Add(item) {
						value = tengo.TrueValue
					}

					return value, nil
				},
			},
			"add_range": &tengo.UserFunction{
				Name: "add",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}

					itemsArray, ok := args[0].(*tengo.Array)
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "items",
							Expected: "array",
							Found:    args[0].TypeName(),
						}
					}

					items, err := arrayToStringSlice(itemsArray)
					if err != nil {
						return nil, err
					}

					for _, item := range items {
						stringSet.Add(item)
					}

					return nil, nil
				},
			},
			"sorted_string_slice": &tengo.UserFunction{
				Name:  "sorted_string_slice",
				Value: stdlib.FuncARSs(stringSet.SortedStringSlice),
			},
		},
	}, nil
}
