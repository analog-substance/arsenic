package engine

import (
	"math/rand"
	"sort"
	"time"

	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) SliceModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"sort_strings":    &tengo.UserFunction{Name: "sort_strings", Value: s.sortStrings},
		"contains_string": &tengo.UserFunction{Name: "contains_string", Value: s.containsString},
		"rand_item":       &tengo.UserFunction{Name: "rand_item", Value: s.randItem},
		"unique":          &tengo.UserFunction{Name: "unique", Value: s.unique},
	}
}

func (s *Script) sortStrings(args ...tengo.Object) (tengo.Object, error) {
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

	slice, err := arrayToStringSlice(array)
	if err != nil {
		return nil, err
	}

	sort.Strings(slice)

	return sliceToStringArray(slice), nil
}

func (s *Script) randItem(args ...tengo.Object) (tengo.Object, error) {
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

	slice := arrayToSlice(array)
	if len(slice) == 0 {
		return nil, nil
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	i := r1.Intn(len(slice))

	return slice[i].(tengo.Object), nil
}

func (s *Script) unique(args ...tengo.Object) (tengo.Object, error) {
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

	slice, err := arrayToStringSlice(array)
	if err != nil {
		return nil, err
	}

	itemSet := set.NewStringSet(slice)
	return sliceToStringArray(itemSet.SortedStringSlice()), nil
}

func (s *Script) containsString(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
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

	input, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "input",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	slice, err := arrayToStringSlice(array)
	if err != nil {
		return nil, err
	}

	for _, item := range slice {
		if item == input {
			return tengo.TrueValue, nil
		}
	}
	return tengo.FalseValue, nil
}
