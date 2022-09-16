package engine

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"

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

func toError(err error) tengo.Object {
	return &tengo.Error{
		Value: &tengo.String{
			Value: err.Error(),
		},
	}
}

func runWithError(cmd *exec.Cmd) error {
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	errorText := ""
	go func() {
		buf := new(bytes.Buffer)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			buf.WriteString(scanner.Text() + "\n")
		}

		errorText = buf.String()
	}()

	err := cmd.Wait()
	if err != nil {
		return errors.New(errorText)
	}
	return nil
}
