package script

import (
	"os"
	"path/filepath"

	"github.com/d5/tengo/v2"
)

var filepathModule *FilePathModule = &FilePathModule{}

type FilePathModule struct {
	moduleMap map[string]tengo.Object
}

func (m *FilePathModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"join":   &tengo.UserFunction{Name: "join", Value: m.join},
			"exists": &tengo.UserFunction{Name: "exists", Value: m.exists},
			"base":   &tengo.UserFunction{Name: "base", Value: m.base},
			"abs":    &tengo.UserFunction{Name: "abs", Value: m.abs},
		}
	}
	return m.moduleMap
}

func (m *FilePathModule) join(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	var paths []string
	for _, obj := range args {
		path, _ := tengo.ToString(obj)
		paths = append(paths, path)
	}

	return &tengo.String{Value: filepath.Join(paths...)}, nil
}

func (m *FilePathModule) exists(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	_, err := os.Stat(path)
	obj := tengo.TrueValue
	if os.IsNotExist(err) {
		obj = tengo.FalseValue
	}

	return obj, nil
}

func (m *FilePathModule) base(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	return &tengo.String{Value: filepath.Base(path)}, nil
}

func (m *FilePathModule) abs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return &tengo.String{Value: absPath}, nil
}
