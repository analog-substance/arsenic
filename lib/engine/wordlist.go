package engine

import (
	"os"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

func (s *Script) WordlistModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"generate": &tengo.UserFunction{
			Name:  "generate",
			Value: s.generateWordlist,
		},
		"types": &tengo.UserFunction{
			Name:  "types",
			Value: stdlib.FuncARSs(lib.GetValidWordlistTypes),
		},
	}
}

func (s *Script) generateWordlist(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	wordlist, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "wordlist",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	path, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	wordlistSet := set.NewStringSet()
	lib.GenerateWordlist(wordlist, wordlistSet)

	file, err := os.Create(path)
	if err != nil {
		return toError(err), nil
	}
	defer file.Close()
	wordlistSet.WriteSorted(file)

	return nil, nil
}
