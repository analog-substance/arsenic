package engine

import (
	"os"

	"github.com/analog-substance/arsenic/pkg"
	"github.com/analog-substance/arsenic/pkg/set"
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
			Value: stdlib.FuncARSs(pkg.GetValidWordlistTypes),
		},
	}
}

func (s *Script) generateWordlist(args interop.ArgMap) (tengo.Object, error) {
	wordlist, _ := args.GetString("wordlist")
	path, _ := args.GetString("path")

	wordlistSet := set.NewStringSet()
	pkg.GenerateWordlist(wordlist, wordlistSet)

	file, err := os.Create(path)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}
	defer file.Close()
	wordlistSet.WriteSorted(file)

	return nil, nil
}
