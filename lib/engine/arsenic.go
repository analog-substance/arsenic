package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	modexec "github.com/analog-substance/tengomod/exec"
	"github.com/analog-substance/tengomod/interop"
)

func (s *Script) ArsenicModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"host": &interop.AdvFunction{
			Name:    "host",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("hostname")},
			Value:   s.host,
		},
		"hosts": &interop.AdvFunction{
			Name:    "hosts",
			NumArgs: interop.MaxArgs(1),
			Args:    []interop.AdvArg{interop.StrSliceArg("flags", false)},
			Value:   s.hosts,
		},
		"locked_files": &interop.AdvFunction{
			Name:    "locked_files",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("glob")},
			Value:   s.lockedFiles,
		},
		"ffuf": &interop.AdvFunction{
			Name:    "ffuf",
			NumArgs: interop.MinArgs(1),
			Args:    []interop.AdvArg{interop.StrSliceArg("args", true)},
			Value:   s.ffuf,
		},
	}
}

func (s *Script) host(args interop.ArgMap) (tengo.Object, error) {
	hostname, _ := args.GetString("hostname")

	foundHost := host.GetFirst(hostname)
	if foundHost == nil {
		return nil, nil
	}

	return makeArsenicHost(foundHost), nil
}

func (s *Script) hosts(args interop.ArgMap) (tengo.Object, error) {
	flags, _ := args.GetStringSlice("flags")

	var hosts []tengo.Object
	for _, h := range host.All() {
		if len(flags) > 0 && !h.Metadata.HasAllFlags(flags...) {
			continue
		}
		hosts = append(hosts, makeArsenicHost(h))
	}

	return &tengo.ImmutableArray{Value: hosts}, nil
}

func (s *Script) lockedFiles(args interop.ArgMap) (tengo.Object, error) {
	glob, _ := args.GetString("glob")

	matches, err := filepath.Glob(glob)
	if err != nil {
		return interop.GoErrToTErr(err), nil
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

	return interop.GoStrSliceToTArray(locked), nil
}

func (s *Script) ffuf(args interop.ArgMap) (tengo.Object, error) {
	cmdArgs, _ := args.GetStringSlice("args")

	cmd := exec.CommandContext(context.Background(), "as-ffuf", cmdArgs...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	errBuf := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(errBuf, os.Stderr)

	err := modexec.RunCmdWithSigHandler(cmd)
	if err != nil {
		return interop.GoErrToTErr(err), nil
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
