package engine

import (
	"net/url"

	"github.com/d5/tengo/v2"
)

func (s *Script) URLModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"hostname": &tengo.UserFunction{Name: "hostname", Value: s.hostname},
	}
}

func (s *Script) hostname(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	rawURL, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "url",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{Value: parsedURL.Hostname()}, nil
}
