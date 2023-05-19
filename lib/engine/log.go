package engine

import (
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengomod/interop"
)

func (s *Script) LogModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"msg":  &tengo.UserFunction{Name: "msg", Value: s.logMsg},
		"warn": &tengo.UserFunction{Name: "warn", Value: s.logWarn},
		"info": &tengo.UserFunction{Name: "info", Value: s.logInfo},
	}
}

func (s *Script) logMsg(args ...tengo.Object) (tengo.Object, error) {
	err := s.log("[+]", args...)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return nil, nil
}

func (s *Script) logWarn(args ...tengo.Object) (tengo.Object, error) {
	err := s.log("[!]", args...)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return nil, nil
}

func (s *Script) logInfo(args ...tengo.Object) (tengo.Object, error) {
	err := s.log("[-]", args...)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return nil, nil
}

func (s *Script) log(prefix string, args ...tengo.Object) error {
	logArgs, err := s.getLogArgs(args...)
	if err != nil {
		return err
	}

	util.Log(prefix, logArgs...)

	return nil
}

func (s *Script) getLogArgs(args ...tengo.Object) ([]interface{}, error) {
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
