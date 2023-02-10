package engine

import (
	"bufio"
	"os"
	"regexp"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	"github.com/andrew-d/go-termutil"
)

func (s *Script) OS2ModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"write_file":         &tengo.UserFunction{Name: "write_file", Value: s.writeFile},
		"read_file_lines":    &tengo.UserFunction{Name: "read_file_lines", Value: s.readFileLines},
		"regex_replace_file": &tengo.UserFunction{Name: "regex_replace_file", Value: s.regexReplaceFile},
		"mkdir_temp":         &tengo.UserFunction{Name: "mkdir_temp", Value: s.mkdirTemp},
		"read_stdin":         &tengo.UserFunction{Name: "read_stdin", Value: s.readStdin},
		"temp_chdir":         &tengo.UserFunction{Name: "temp_chdir", Value: s.tempChdir},
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

func (s *Script) readFileLines(args ...tengo.Object) (tengo.Object, error) {
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

	lines, err := util.ReadLines(path)
	if err != nil {
		return toError(err), nil
	}

	return sliceToStringArray(lines), nil
}

func (s *Script) readStdin(args ...tengo.Object) (tengo.Object, error) {
	if termutil.Isatty(os.Stdin.Fd()) {
		return nil, nil
	}

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return sliceToStringArray(lines), nil
}

func (s *Script) tempChdir(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
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

	fn, ok := args[1].(*tengo.CompiledFunction)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "fn",
			Expected: "function",
			Found:    args[1].TypeName(),
		}
	}

	var err error
	previousDir := ""

	if path != "" {
		previousDir, err = os.Getwd()
		if err != nil {
			return toError(err), nil
		}

		err = os.Chdir(path)
		if err != nil {
			return toError(err), nil
		}
	}

	err = s.runCompiledFunction(fn)
	if err != nil {
		return toError(err), nil
	}

	if path != "" {
		err = os.Chdir(previousDir)
		if err != nil {
			return toError(err), nil
		}
	}

	return nil, nil
}
