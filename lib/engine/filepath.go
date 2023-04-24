package engine

import (
	"path/filepath"
	"regexp"

	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/bmatcuk/doublestar/v4"
)

func (s *Script) FilePathModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"join":        &tengo.UserFunction{Name: "join", Value: funcASvRS(filepath.Join)},
		"file_exists": &tengo.UserFunction{Name: "file_exists", Value: funcASRB(fileutil.FileExists)},
		"dir_exists":  &tengo.UserFunction{Name: "dir_exists", Value: funcASRB(fileutil.DirExists)},
		"base":        &tengo.UserFunction{Name: "base", Value: stdlib.FuncASRS(filepath.Base)},
		"dir":         &tengo.UserFunction{Name: "dir", Value: stdlib.FuncASRS(filepath.Dir)},
		"abs":         &tengo.UserFunction{Name: "abs", Value: stdlib.FuncASRSE(filepath.Abs)},
		"ext":         &tengo.UserFunction{Name: "ext", Value: stdlib.FuncASRS(filepath.Ext)},
		"glob":        &tengo.UserFunction{Name: "glob", Value: s.glob},
		"from_slash":  &tengo.UserFunction{Name: "from_slash", Value: stdlib.FuncASRS(filepath.FromSlash)},
	}
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
