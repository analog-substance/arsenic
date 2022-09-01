package script

import (
	"github.com/d5/tengo/v2"
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

func toStringArray(slice []string) tengo.Object {
	var values []tengo.Object
	for _, s := range slice {
		values = append(values, &tengo.String{Value: s})
	}

	return &tengo.Array{
		Value: values,
	}
}
