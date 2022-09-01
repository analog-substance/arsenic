package script

import (
	"os"
	"strings"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/d5/tengo/v2"
)

var arsenicModule *ArsenicModule = &ArsenicModule{}

type ArsenicModule struct {
	moduleMap map[string]tengo.Object
}

func (m *ArsenicModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"host_urls":    &tengo.UserFunction{Name: "host_urls", Value: m.hostUrls},
			"host_path":    &tengo.UserFunction{Name: "host_path", Value: m.hostPath},
			"gen_wordlist": &tengo.UserFunction{Name: "gen_wordlist", Value: m.generateWordlist},
		}
	}
	return m.moduleMap
}

func (m *ArsenicModule) hostUrls(args ...tengo.Object) (tengo.Object, error) {
	var protocols []string
	for _, arg := range args {
		protocol, ok := tengo.ToString(arg)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "protocol",
				Expected: "string",
				Found:    arg.TypeName(),
			}
		}

		protocols = append(protocols, protocol)
	}

	hosts := host.All()
	hostURLs := set.NewStringSet()
	for _, h := range hosts {
		hostURLs.AddRange(h.URLs())
	}

	validHostURLs := set.NewStringSet()
	for _, hostURL := range hostURLs.SortedStringSlice() {
		for _, proto := range protocols {
			if strings.HasPrefix(hostURL, proto) || proto == "all" {
				validHostURLs.Add(hostURL)
			}
		}
	}

	return toStringArray(validHostURLs.SortedStringSlice()), nil
}

func (m *ArsenicModule) hostPath(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	hostname, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "hostname",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	hosts := host.Get(hostname)

	var paths []string
	for _, h := range hosts {
		paths = append(paths, h.Dir)
	}

	return toStringArray(paths), nil
}

func (m *ArsenicModule) generateWordlist(args ...tengo.Object) (tengo.Object, error) {
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

	wordlistSet := set.NewSet("")
	lib.GenerateWordlist(wordlist, &wordlistSet)

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	wordlistSet.WriteSorted(file)

	return nil, nil
}
