package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/d5/tengo/v2"
)

func (s *Script) ArsenicModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"host_urls":    &tengo.UserFunction{Name: "host_urls", Value: s.hostUrls},
		"host_path":    &tengo.UserFunction{Name: "host_path", Value: s.hostPath},
		"host_paths":   &tengo.UserFunction{Name: "host_paths", Value: s.hostPaths},
		"gen_wordlist": &tengo.UserFunction{Name: "gen_wordlist", Value: s.generateWordlist},
		"locked_files": &tengo.UserFunction{Name: "locked_files", Value: s.lockedFiles},
		"ffuf":         &tengo.UserFunction{Name: "ffuf", Value: s.ffuf},
		"tcp_scan":     &tengo.UserFunction{Name: "tcp_scan", Value: s.tcpScan},
	}
}

func (s *Script) hostUrls(args ...tengo.Object) (tengo.Object, error) {
	var protocols []string
	for _, arg := range args {
		protocol, ok := tengo.ToString(arg)
		if !ok {
			return toError(tengo.ErrInvalidArgumentType{
				Name:     "protocol",
				Expected: "string",
				Found:    arg.TypeName(),
			}), nil
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

func (s *Script) hostPath(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	hostname, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "hostname",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	foundHost := host.GetFirst(hostname)
	if foundHost == nil {
		return nil, nil
	}

	return &tengo.String{Value: foundHost.Dir}, nil
}

func (s *Script) hostPaths(args ...tengo.Object) (tengo.Object, error) {
	var paths []string
	hosts := host.All()
	for _, h := range hosts {
		paths = append(paths, h.Dir)
	}

	return toStringArray(paths), nil
}

func (s *Script) generateWordlist(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	wordlist, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "wordlist",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	path, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
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

func (s *Script) lockedFiles(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	glob, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "glob",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	matches, err := filepath.Glob(glob)
	if err != nil {
		return toError(err), nil
	}

	lockRegex := regexp.MustCompile(`^lock::`)

	var locked []string
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			util.LogWarn(err)
			continue
		}

		if lockRegex.Match(data) {
			locked = append(locked, match)
		}
	}

	return toStringArray(locked), nil
}

func (s *Script) ffuf(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	var cmdArgs []string
	for _, arg := range args {
		cmdArg, ok := tengo.ToString(arg)
		if !ok {
			return toError(tengo.ErrInvalidArgumentType{
				Name:     "ffuf arg",
				Expected: "string",
				Found:    arg.TypeName(),
			}), nil
		}

		cmdArgs = append(cmdArgs, cmdArg)
	}

	err := s.runWithSigHandler("as-ffuf", cmdArgs...)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}

func (s *Script) tcpScan(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	var cmdArgs []string
	for _, arg := range args {
		cmdArg, ok := tengo.ToString(arg)
		if !ok {
			return toError(tengo.ErrInvalidArgumentType{
				Name:     "as-recon-discover-service arg",
				Expected: "string",
				Found:    arg.TypeName(),
			}), nil
		}

		cmdArgs = append(cmdArgs, cmdArg)
	}

	err := s.runWithSigHandler("as-recon-discover-services", cmdArgs...)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}
