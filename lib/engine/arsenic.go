package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) ArsenicModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"host_urls":              &tengo.UserFunction{Name: "host_urls", Value: s.hostUrls},
		"host":                   &tengo.UserFunction{Name: "host", Value: s.host},
		"hosts":                  &tengo.UserFunction{Name: "hosts", Value: s.hosts},
		"gen_wordlist":           &tengo.UserFunction{Name: "gen_wordlist", Value: s.generateWordlist},
		"locked_files":           &tengo.UserFunction{Name: "locked_files", Value: s.lockedFiles},
		"ffuf":                   &tengo.UserFunction{Name: "ffuf", Value: s.ffuf},
		"content_discovery_urls": &tengo.UserFunction{Name: "content_discovery_urls", Value: s.contentDiscoveryURLs},
	}
}

func (s *Script) hostUrls(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	protocolsArray, ok := args[0].(*tengo.Array)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "protocols",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	protocols, err := arrayToStringSlice(protocolsArray)
	if err != nil {
		return nil, err
	}

	var flags []string
	if len(args) == 2 {
		flagsArray, ok := args[1].(*tengo.Array)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "flags",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}

		flags, err = arrayToStringSlice(flagsArray)
		if err != nil {
			return nil, err
		}
	}

	hosts := host.All()
	hostURLs := set.NewStringSet()
	for _, h := range hosts {
		if len(flags) > 0 && !h.Metadata.HasFlags(flags...) {
			continue
		}
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

	return sliceToStringArray(validHostURLs.SortedStringSlice()), nil
}

func (s *Script) host(args ...tengo.Object) (tengo.Object, error) {
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

	foundHost := host.GetFirst(hostname)
	if foundHost == nil {
		return nil, nil
	}

	return makeArsenicHost(foundHost), nil
}

func (s *Script) hosts(args ...tengo.Object) (tengo.Object, error) {
	var flags []string
	if len(args) == 1 {
		flagsArray, ok := args[0].(*tengo.Array)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "flags",
				Expected: "string",
				Found:    args[0].TypeName(),
			}
		}

		var err error
		flags, err = arrayToStringSlice(flagsArray)
		if err != nil {
			return nil, err
		}
	}

	var hosts []tengo.Object
	for _, h := range host.All() {
		if len(flags) > 0 && !h.Metadata.HasAllFlags(flags...) {
			continue
		}
		hosts = append(hosts, makeArsenicHost(h))
	}

	return &tengo.ImmutableArray{Value: hosts}, nil
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

func (s *Script) lockedFiles(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	glob, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "glob",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
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

	return sliceToStringArray(locked), nil
}

func (s *Script) ffuf(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	cmdArgs, err := sliceToStringSlice(args)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(context.Background(), "as-ffuf", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	errBuf := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(errBuf, os.Stderr)

	err = s.runCmdWithSigHandler(cmd)
	if err != nil {
		return toError(err), nil
	}

	warnRe := regexp.MustCompile(`(?m)\[WARN\]\s*(.*)$`)
	matches := warnRe.FindAllStringSubmatch(errBuf.String(), -1)
	if len(matches) != 0 {
		var warnings []tengo.Object
		for _, match := range matches {
			warnings = append(warnings, toWarning(match[1]))
		}

		return &tengo.Array{Value: warnings}, nil
	}

	return nil, nil
}

func (s *Script) contentDiscoveryURLs(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	patternsArray, ok := args[0].(*tengo.Array)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "patterns",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	patterns, err := arrayToStringSlice(patternsArray)
	if err != nil {
		return nil, err
	}

	codesArray, ok := args[1].(*tengo.Array)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "codes",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	codes, err := arrayToIntSlice(codesArray)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return toError(err), nil
		}
		files = append(files, matches...)
	}

	allResults, err := gocdp.SmartParseFiles(files)
	if err != nil {
		return toError(err), nil
	}
	grouped := allResults.GroupByStatus()

	var urls []string
	for _, code := range codes {
		results, ok := grouped[code]
		if !ok {
			continue
		}

		for _, result := range results {
			urls = append(urls, result.Url)
		}
	}

	return sliceToStringArray(urls), nil
}
