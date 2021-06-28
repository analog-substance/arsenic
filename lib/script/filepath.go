package script

import (
	"path/filepath"

	"github.com/d5/tengo/v2"
)

var filepathModule = map[string]tengo.Object{
	"join": &tengo.UserFunction{Name: "join", Value: filepathJoin},
}

func filepathJoin(args ...tengo.Object) (tengo.Object, error) {
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
