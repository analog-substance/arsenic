package engine

import (
	"path/filepath"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	"github.com/bmatcuk/doublestar/v4"
)

func (m *Script) FilePathModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"join":   &tengo.UserFunction{Name: "join", Value: m.join},
		"exists": &tengo.UserFunction{Name: "exists", Value: m.exists},
		"base":   &tengo.UserFunction{Name: "base", Value: m.base},
		"abs":    &tengo.UserFunction{Name: "abs", Value: m.abs},
		"ext":    &tengo.UserFunction{Name: "ext", Value: m.ext},
		"glob":   &tengo.UserFunction{Name: "glob", Value: m.glob},
	}
}

func (m *Script) join(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	var paths []string
	for _, obj := range args {
		path, _ := tengo.ToString(obj)
		paths = append(paths, path)
	}

	return &tengo.String{Value: filepath.Join(paths...)}, nil
}

func (m *Script) exists(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
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

	obj := tengo.TrueValue
	if !util.FileExists(path) {
		obj = tengo.FalseValue
	}

	return obj, nil
}

func (m *Script) base(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
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

	return &tengo.String{Value: filepath.Base(path)}, nil
}

func (m *Script) abs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
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

	absPath, err := filepath.Abs(path)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{Value: absPath}, nil
}

func (m *Script) ext(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
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

	return &tengo.String{Value: filepath.Ext(path)}, nil
}

func (m *Script) glob(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	pattern, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "pattern",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return toError(err), nil
	}

	return toStringArray(matches), nil
}
