package engine

import (
	"path/filepath"
	"regexp"

	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/bmatcuk/doublestar/v4"
)

func (s *Script) FilePathModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"join":        &tengo.UserFunction{Name: "join", Value: s.join},
		"file_exists": &tengo.UserFunction{Name: "file_exists", Value: s.fileExists},
		"dir_exists":  &tengo.UserFunction{Name: "dir_exists", Value: s.dirExists},
		"base":        &tengo.UserFunction{Name: "base", Value: s.base},
		"abs":         &tengo.UserFunction{Name: "abs", Value: s.abs},
		"ext":         &tengo.UserFunction{Name: "ext", Value: s.ext},
		"glob":        &tengo.UserFunction{Name: "glob", Value: s.glob},
		"from_slash":  &tengo.UserFunction{Name: "from_slash", Value: s.fromSlash},
	}
}

func (s *Script) join(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	paths, err := sliceToStringSlice(args)
	if err != nil {
		return nil, err
	}

	return &tengo.String{Value: filepath.Join(paths...)}, nil
}

func (s *Script) fileExists(args ...tengo.Object) (tengo.Object, error) {
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

	obj := tengo.TrueValue
	if !fileutil.FileExists(path) {
		obj = tengo.FalseValue
	}

	return obj, nil
}

func (s *Script) dirExists(args ...tengo.Object) (tengo.Object, error) {
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

	obj := tengo.TrueValue
	if !fileutil.DirExists(path) {
		obj = tengo.FalseValue
	}

	return obj, nil
}

func (s *Script) base(args ...tengo.Object) (tengo.Object, error) {
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

func (s *Script) abs(args ...tengo.Object) (tengo.Object, error) {
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
		return toError(err), nil
	}

	return &tengo.String{Value: absPath}, nil
}

func (s *Script) ext(args ...tengo.Object) (tengo.Object, error) {
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

	return &tengo.String{Value: filepath.Ext(path)}, nil
}

func (s *Script) glob(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 || len(args) > 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	pattern, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "pattern",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	var excludeRe *regexp.Regexp
	if len(args) == 2 {
		excludePatternArg, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "exclude-pattern",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		var err error
		excludeRe, err = regexp.Compile(excludePatternArg)
		if err != nil {
			return nil, err
		}
	}

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return toError(err), nil
	}

	if excludeRe != nil {
		var filtered []string
		for _, match := range matches {
			if !excludeRe.MatchString(match) {
				filtered = append(filtered, match)
			}
		}
		return sliceToStringArray(filtered), nil
	}

	return sliceToStringArray(matches), nil
}

func (s *Script) fromSlash(args ...tengo.Object) (tengo.Object, error) {
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

	return &tengo.String{Value: filepath.FromSlash(path)}, nil
}
