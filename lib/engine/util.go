package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"github.com/analog-substance/tengo/v2"
)

func toStringSlice(array *tengo.Array) ([]string, error) {
	var slice []string
	for _, v := range array.Value {
		str, ok := tengo.ToString(v)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "Array type",
				Expected: "string",
				Found:    v.TypeName(),
			}
		}

		slice = append(slice, str)
	}

	return slice, nil
}

func toSlice(array *tengo.Array) []interface{} {
	var slice []interface{}
	for _, v := range array.Value {
		slice = append(slice, v)
	}

	return slice
}

func toIntSlice(array *tengo.Array) ([]int, error) {
	var slice []int
	for _, v := range array.Value {
		i, ok := tengo.ToInt(v)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "Array type",
				Expected: "int",
				Found:    v.TypeName(),
			}
		}

		slice = append(slice, i)
	}

	return slice, nil
}

func toStringArray(slice []string) tengo.Object {
	var values []tengo.Object
	for _, s := range slice {
		values = append(values, &tengo.String{Value: s})
	}

	return &tengo.Array{
		Value: values,
	}
}

func toIntArray(slice []int) tengo.Object {
	var values []tengo.Object
	for _, i := range slice {
		values = append(values, &tengo.Int{Value: int64(i)})
	}

	return &tengo.Array{
		Value: values,
	}
}

func toError(err error) tengo.Object {
	return &tengo.Error{
		Value: &tengo.String{
			Value: err.Error(),
		},
	}
}

func toWarning(value string) tengo.Object {
	return &Warning{
		Value: &tengo.String{
			Value: value,
		},
	}
}

func runWithError(cmd *exec.Cmd) error {
	buf := new(bytes.Buffer)
	cmd.Stderr = buf

	if err := cmd.Start(); err != nil {
		return err
	}

	err := cmd.Wait()
	if err != nil {
		return errors.New(buf.String())
	}
	return nil
}

func funcASSSSRSp(fn func(string, string, string, string) *string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}
		s3, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}
		s4, ok := tengo.ToString(args[3])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "fourth",
				Expected: "string(compatible)",
				Found:    args[3].TypeName(),
			}
		}
		s := fn(s1, s2, s3, s4)
		if len(*s) > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		return &tengo.String{Value: *s}, nil
	}
}

func funcASSSRSp(fn func(string, string, string) *string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}
		s3, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}
		s := fn(s1, s2, s3)
		if len(*s) > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		return &tengo.String{Value: *s}, nil
	}
}

func funcASSSsSRSsp(fn func(string, string, []string, string) *[]string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		var ss1 []string
		switch arg2 := args[2].(type) {
		case *tengo.Array:
			for idx, a := range arg2.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("third[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		case *tengo.ImmutableArray:
			for idx, a := range arg2.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("third[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "array",
				Found:    args[0].TypeName(),
			}
		}

		s4, ok := tengo.ToString(args[3])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "fourth",
				Expected: "string(compatible)",
				Found:    args[3].TypeName(),
			}
		}

		arr := &tengo.Array{}
		for _, res := range *fn(s1, s2, ss1, s4) {
			if len(res) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: res})
		}
		return arr, nil
	}
}

func funcASSsSRSsp(fn func(string, []string, string) *[]string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		var ss1 []string
		switch arg1 := args[1].(type) {
		case *tengo.Array:
			for idx, a := range arg1.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("second[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		case *tengo.ImmutableArray:
			for idx, a := range arg1.Value {
				as, ok := tengo.ToString(a)
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     fmt.Sprintf("second[%d]", idx),
						Expected: "string(compatible)",
						Found:    a.TypeName(),
					}
				}
				ss1 = append(ss1, as)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "array",
				Found:    args[1].TypeName(),
			}
		}

		s3, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}

		arr := &tengo.Array{}
		for _, res := range *fn(s1, ss1, s3) {
			if len(res) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: res})
		}
		return arr, nil
	}
}

func funcASSBSRBp(fn func(string, string, bool, string) *bool) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		b1, ok := tengo.ToBool(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "bool(compatible)",
				Found:    args[2].TypeName(),
			}
		}

		s4, ok := tengo.ToString(args[3])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "fourth",
				Expected: "string(compatible)",
				Found:    args[3].TypeName(),
			}
		}

		if *fn(s1, s2, b1, s4) {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}

func funcASBSRBp(fn func(string, bool, string) *bool) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		b1, ok := tengo.ToBool(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "bool(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s4, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}

		if *fn(s1, b1, s4) {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}

func funcASSISRIp(fn func(string, string, int, string) *int) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		s2, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "string(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		i1, ok := tengo.ToInt(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "int(compatible)",
				Found:    args[2].TypeName(),
			}
		}

		s4, ok := tengo.ToString(args[3])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "fourth",
				Expected: "string(compatible)",
				Found:    args[3].TypeName(),
			}
		}

		i := fn(s1, s2, i1, s4)
		return &tengo.Int{Value: int64(*i)}, nil
	}
}

func funcASISRIp(fn func(string, int, string) *int) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 3 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		i1, ok := tengo.ToInt(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "second",
				Expected: "int(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s4, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "third",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}

		i := fn(s1, i1, s4)
		return &tengo.Int{Value: int64(*i)}, nil
	}
}

func funcASRSsE(fn func(string) ([]string, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		res, err := fn(s1)
		if err != nil {
			return toError(err), nil
		}

		arr := &tengo.Array{}
		for _, r := range res {
			if len(r) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: r})
		}
		return arr, nil
	}
}

func funcASRBE(fn func(string) (bool, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		res, err := fn(s1)
		if err != nil {
			return toError(err), nil
		}

		if res {
			return tengo.TrueValue, nil
		}
		return tengo.FalseValue, nil
	}
}
