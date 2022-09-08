package script

import (
	"net/url"

	"github.com/d5/tengo/v2"
)

var urlModule *URLModule = &URLModule{}

type URLModule struct {
	moduleMap map[string]tengo.Object
}

func (m *URLModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"hostname": &tengo.UserFunction{Name: "hostname", Value: m.hostname},
		}
	}
	return m.moduleMap
}

func (m *URLModule) hostname(args ...tengo.Object) (tengo.Object, error) {
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
