package engine

import (
	"math/rand"
	"sort"
	"time"

	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) SliceModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"sort_strings": &tengo.UserFunction{Name: "sort_strings", Value: s.sortStrings},
		"rand_item":    &tengo.UserFunction{Name: "rand_item", Value: s.randItem},
		"unique":       &tengo.UserFunction{Name: "unique", Value: s.unique},
	}
}

func (s *Script) sortStrings(args ...tengo.Object) (tengo.Object, error) {
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

	slice, err := arrayToStringSlice(array)
	if err != nil {
		return toError(err), nil
	}

	sort.Strings(slice)

	return sliceToStringArray(slice), nil
}

func (s *Script) randItem(args ...tengo.Object) (tengo.Object, error) {
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

	slice, err := arrayToStringSlice(array)
	if err != nil {
		return toError(err), nil
	}

	itemSet := set.NewStringSet(slice)
	return sliceToStringArray(itemSet.SortedStringSlice()), nil
}
