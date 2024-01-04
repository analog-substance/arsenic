package engine

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/log"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengomod/interop"
)

func (s *Script) ArsenicModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"add_host": &interop.AdvFunction{
			Name:    "add_host",
			NumArgs: interop.ExactArgs(2),
			Args: []interop.AdvArg{
				interop.StrSliceArg("hostnames", false),
				interop.StrSliceArg("ips", false),
			},
			Value: s.addHost,
		},
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
			log.Warn(err)
			continue
		}

		if lockRegex.Match(data) {
			locked = append(locked, match)
		}
	}

	return interop.GoStrSliceToTArray(locked), nil
}

func (s *Script) addHost(args interop.ArgMap) (tengo.Object, error) {
	hostnames, _ := args.GetStringSlice("hostnames")
	ips, _ := args.GetStringSlice("ips")

	h, err := host.AddHost(hostnames, ips)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	if h == nil {
		return interop.GoErrToTErr(errors.New("must supply hostnames and ips")), nil
	}

	return makeArsenicHost(h), nil
}
