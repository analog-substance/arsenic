package engine

import (
	"os"
	"regexp"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) OS2ModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"write_file":         &tengo.UserFunction{Name: "write_file", Value: s.writeFile},
		"regex_replace_file": &tengo.UserFunction{Name: "regex_replace_file", Value: s.regexReplaceFile},
		"mkdir_temp":         &tengo.UserFunction{Name: "mkdir_temp", Value: s.mkdirTemp},
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

func (s *Script) regexReplaceFile(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 3 {
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

	regex, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "regex",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return toError(err), nil
	}

	replace, ok := tengo.ToString(args[2])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "replace",
			Expected: "string",
			Found:    args[2].TypeName(),
		}), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return toError(err), nil
	}

	replaced := re.ReplaceAll(data, []byte(replace))

	err = os.WriteFile(path, replaced, util.DefaultFilePerms)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

func (s *Script) mkdirTemp(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	dir, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "dir",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	pattern, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "pattern",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
	}

	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{
		Value: tempDir,
	}, nil
}
