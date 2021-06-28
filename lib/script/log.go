package script

import (
	"fmt"

	"github.com/d5/tengo/v2"
)

var logModule = map[string]tengo.Object{
	"log":  &tengo.UserFunction{Name: "log", Value: logLog},
	"warn": &tengo.UserFunction{Name: "warn", Value: logWarn},
	"info": &tengo.UserFunction{Name: "info", Value: logInfo},
}

func logLog(args ...tengo.Object) (tengo.Object, error) {
	err := log("[+]", args...)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func logWarn(args ...tengo.Object) (tengo.Object, error) {
	err := log("[!]", args...)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func logInfo(args ...tengo.Object) (tengo.Object, error) {
	err := log("[-]", args...)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func log(prefix string, args ...tengo.Object) error {
	logArgs, err := getLogArgs(args...)
	if err != nil {
		return err
	}

	fmt.Printf("%s ", prefix)
	fmt.Print(logArgs...)
	fmt.Println()

	return nil
}

func getLogArgs(args ...tengo.Object) ([]interface{}, error) {
	var logArgs []interface{}
	l := 0
	for _, arg := range args {
		s, _ := tengo.ToString(arg)
		slen := len(s)
		// make sure length does not exceed the limit
		if l+slen > tengo.MaxStringLen {
			return nil, tengo.ErrStringLimit
		}
		l += slen
		logArgs = append(logArgs, s)
	}
	return logArgs, nil
}
