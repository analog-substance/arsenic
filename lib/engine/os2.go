package engine

import (
	"os"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/d5/tengo/v2"
)

func (s *Script) OS2ModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"write_file": &tengo.UserFunction{Name: "write_file", Value: s.writeFile},
	}
}

func (s *Script) writeFile(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	data, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "data",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
	}

	err := os.WriteFile(path, []byte(data), util.DefaultFilePerms)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}
