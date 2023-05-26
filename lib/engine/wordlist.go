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
		"generate": &interop.AdvFunction{
			Name:    "generate",
			NumArgs: interop.ExactArgs(2),
			Args:    []interop.AdvArg{interop.StrArg("wordlist"), interop.StrArg("path")},
			Value:   s.generateWordlist,
		},
		"types": &tengo.UserFunction{
			Name:  "types",
			Value: stdlib.FuncARSs(lib.GetValidWordlistTypes),
		},
	}
}

func (s *Script) generateWordlist(args map[string]interface{}) (tengo.Object, error) {
	wordlist := args["wordlist"].(string)
	path := args["path"].(string)

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
