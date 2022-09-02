package script

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/analog-substance/arsenic/lib"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
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
			"locked_files": &tengo.UserFunction{Name: "locked_files", Value: m.lockedFiles},
			"ffuf":         &tengo.UserFunction{Name: "ffuf", Value: m.ffuf},
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

	foundHost := host.GetFirst(hostname)
	if foundHost == nil {
		return nil, nil
	}

	return &tengo.String{Value: foundHost.Dir}, nil
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

func (m *ArsenicModule) lockedFiles(args ...tengo.Object) (tengo.Object, error) {
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
		return nil, err
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

func (m *ArsenicModule) ffuf(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	var cmdArgs []string
	for _, arg := range args {
		cmdArg, ok := tengo.ToString(arg)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "ffuf arg",
				Expected: "string",
				Found:    arg.TypeName(),
			}
		}

		cmdArgs = append(cmdArgs, cmdArg)
	}

	cmdCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(cmdCtx, "as-ffuf", cmdArgs...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// relay trapped signals to the spawned process
	terminate := false
	go func() {
		for sig := range sigs {
			terminate = true
			cmd.Process.Signal(sig)
			cancel()
		}
	}()

	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					terminate = true
				}
			}
		} else {
			return nil, err
		}
	}

	if terminate {
		stopScript()
	}

	return nil, nil
}
