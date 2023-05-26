package engine

import (
	"os"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod/interop"
)

func (s *Script) WordlistModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"generate": &tengo.UserFunction{
			Name:  "generate",
			Value: interop.NewCallable(s.generateWordlist, interop.WithExactArgs(2)),
		},
		"types": &tengo.UserFunction{
			Name:  "types",
			Value: stdlib.FuncARSs(lib.GetValidWordlistTypes),
		},
	}
}

func (s *Script) generateWordlist(args ...tengo.Object) (tengo.Object, error) {
	wordlist, err := interop.TStrToGoStr(args[0], "wordlist")
	if err != nil {
		return nil, err
	}

	path, err := interop.TStrToGoStr(args[1], "path")
	if err != nil {
		return nil, err
	}

	wordlistSet := set.NewStringSet()
	lib.GenerateWordlist(wordlist, wordlistSet)

	file, err := os.Create(path)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}
	defer file.Close()
	wordlistSet.WriteSorted(file)

	return nil, nil
}
