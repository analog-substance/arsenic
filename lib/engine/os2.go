package engine

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/andrew-d/go-termutil"
	"github.com/bmatcuk/doublestar/v4"
)

// OS2Module represents the 'os2' import module
func (s *Script) OS2Module() map[string]tengo.Object {
	return map[string]tengo.Object{
		"write_file":         &tengo.UserFunction{Name: "write_file", Value: s.writeFile},
		"write_file_lines":   &tengo.UserFunction{Name: "write_file_lines", Value: s.writeFileLines},
		"read_file_lines":    &tengo.UserFunction{Name: "read_file_lines", Value: s.readFileLines},
		"regex_replace_file": &tengo.UserFunction{Name: "regex_replace_file", Value: s.regexReplaceFile},
		"mkdir_all":          &tengo.UserFunction{Name: "mkdir_all", Value: s.mkdirAll},
		"mkdir_temp":         &tengo.UserFunction{Name: "mkdir_temp", Value: s.mkdirTemp},
		"read_stdin":         &tengo.UserFunction{Name: "read_stdin", Value: s.readStdin},
		"temp_chdir":         &tengo.UserFunction{Name: "temp_chdir", Value: s.tempChdir},
		"copy_files":         &tengo.UserFunction{Name: "copy_files", Value: s.copyFiles},
		"copy_dirs":          &tengo.UserFunction{Name: "copy_dirs", Value: s.copyDirs},
	}
}

// writeFile is like the tengo 'os.write_file' function except the file is written with 0644 permissions
// Represents 'os2.write_file(path string, data string) error'
func (s *Script) writeFile(args ...tengo.Object) (tengo.Object, error) {
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

	data, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "data",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	err := fileutil.WriteString(path, data)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

// writeFileLines is like the writeFile function except each element in the slice is written on a new line
// Represents 'os2.write_file_lines(path string, lines []string) error'
func (s *Script) writeFileLines(args ...tengo.Object) (tengo.Object, error) {
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

	lineArray, ok := args[1].(*tengo.Array)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "lines",
			Expected: "[]string",
			Found:    args[1].TypeName(),
		}
	}

	lines, err := arrayToStringSlice(lineArray)
	if err != nil {
		return nil, err
	}

	err = fileutil.WriteLines(path, lines)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

// regexReplaceFile reads the file, replaces the contents that match the regex and writes it back to the file.
// Represents 'os2.regex_replace_file(path string, regex string, replace string) error'
func (s *Script) regexReplaceFile(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 3 {
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

	regex, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "regex",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	replace, ok := tengo.ToString(args[2])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "replace",
			Expected: "string",
			Found:    args[2].TypeName(),
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return toError(err), nil
	}

	replaced := re.ReplaceAll(data, []byte(replace))

	err = fileutil.WriteString(path, string(replaced))
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

// mkdirAll is a simple tengo function wrapper to 'os.MkdirAll' except it sets the directory permissions to 0755
// Represents 'os2.mkdir_all(paths ...string) error'
func (s *Script) mkdirAll(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	for _, obj := range args {
		path, _ := tengo.ToString(obj)
		err := os.MkdirAll(path, fileutil.DefaultDirPerms)
		if err != nil {
			return toError(err), nil
		}
	}

	return nil, nil
}

// mkdirTemp is a tengo function wrapper to the os.MkdirTemp function
// Represents 'os2.mkdir_temp(dir string, pattern string) string|error'
func (s *Script) mkdirTemp(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	dir, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "dir",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	pattern, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "pattern",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	tempDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return toError(err), nil
	}

	return &tengo.String{
		Value: tempDir,
	}, nil
}

// readFileLines reads the file and splits the contents by each new line
// Represents 'os2.read_file_lines(path string) []string|error'
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

	lines, err := fileutil.ReadLines(path)
	if err != nil {
		return toError(err), nil
	}

	return sliceToStringArray(lines), nil
}

// readStdin reads the current process's Stdin if anything was piped to it.
// Represents 'os2.read_stdin() []string'
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

// tempChdir changes the current directory, executes the function, then changes the current directory back.
// Represents 'os2.temp_chdir(dir string, fn func())'
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

	_, err = s.runCompiledFunction(fn)
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

// copyFiles copies the specified files to the destination.
// Represents 'os2.copy_files(src string|[]string, dest string) error'
func (s *Script) copyFiles(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	var err error
	var files []string

	filesArray, ok := args[0].(*tengo.Array)
	if ok {
		files, err = arrayToStringSlice(filesArray)
		if err != nil {
			return nil, err
		}
	} else {
		src, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "src",
				Expected: "string|[]string",
				Found:    args[0].TypeName(),
			}
		}

		files, err = doublestar.FilepathGlob(src)
		if err != nil {
			return toError(err), nil
		}
	}

	dest, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "dest",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	for _, file := range files {
		err = fileutil.CopyFile(file, dest)
		if err != nil {
			return toError(err), nil
		}
	}

	return nil, nil
}

// copyDirs copies the specified directories to the destination.
// Represents 'os2.copy_dirs(src string..., dest string) error'
func (s *Script) copyDirs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	var srcDirs []string
	for _, arg := range args[:len(args)-1] {
		src, ok := tengo.ToString(arg)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "src",
				Expected: "string",
				Found:    arg.TypeName(),
			}
		}

		srcDirs = append(srcDirs, src)
	}

	dest, ok := tengo.ToString(args[len(args)-1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "dest",
			Expected: "string",
			Found:    args[len(args)-1].TypeName(),
		}
	}

	if len(srcDirs) > 1 && !fileutil.DirExists(dest) {
		return toError(fmt.Errorf("%s: No such directory", dest)), nil
	}

	for _, src := range srcDirs {
		err := fileutil.CopyDir(src, dest)
		if err != nil {
			return toError(err), nil
		}
	}

	return nil, nil
}
