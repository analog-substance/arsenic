package engine

import (
	"net/url"

	"github.com/analog-substance/tengo/v2"
)

func (s *Script) URLModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"hostname": &tengo.UserFunction{Name: "hostname", Value: s.hostname},
	}
}

func (s *Script) hostname(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	rawURL, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "url",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{Value: parsedURL.Hostname()}, nil
}
