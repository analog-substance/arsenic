package engine

import (
	"path/filepath"
	"strings"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/d5/tengo/v2"
)

func makeArsenicHost(h *host.Host) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"dir": &tengo.String{
				Value: h.Dir,
			},
			"name": &tengo.String{
				Value: h.Metadata.Name,
			},
			"has_flags": &tengo.UserFunction{
				Name: "has_flags",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					var flags []string
					for _, arg := range args {
						flag, ok := tengo.ToString(arg)
						if !ok {
							return toError(tengo.ErrInvalidArgumentType{
								Name:     "flag",
								Expected: "string",
								Found:    arg.TypeName(),
							}), nil
						}

						flags = append(flags, flag)
					}

					value := tengo.FalseValue
					if h.Metadata.HasFlags(flags...) {
						value = tengo.TrueValue
					}

					return value, nil
				},
			},
			"files": &tengo.UserFunction{
				Name: "files",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					var globs []string
					for _, arg := range args {
						glob, ok := tengo.ToString(arg)
						if !ok {
							return toError(tengo.ErrInvalidArgumentType{
								Name:     "glob",
								Expected: "string",
								Found:    arg.TypeName(),
							}), nil
						}

						globs = append(globs, glob)
					}

					matches, err := h.Files(globs...)
					if err != nil {
						return toError(err), nil
					}
					return toStringArray(matches), nil
				},
			},
			"urls": &tengo.UserFunction{
				Name: "urls",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
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
					if len(protocols) == 0 {
						protocols = append(protocols, "all")
					}

					var urls []string
					for _, hostURL := range h.URLs() {
						for _, proto := range protocols {
							if strings.HasPrefix(hostURL, proto) || proto == "all" {
								urls = append(urls, hostURL)
							}
						}
					}

					return toStringArray(urls), nil
				},
			},
			"file_exists": &tengo.UserFunction{
				Name: "file_exists",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 1 {
						return toError(tengo.ErrWrongNumArguments), nil
					}

					file, ok := tengo.ToString(args[0])
					if !ok {
						return toError(tengo.ErrInvalidArgumentType{
							Name:     "path",
							Expected: "string",
							Found:    args[0].TypeName(),
						}), nil
					}

					value := tengo.FalseValue
					if util.FileExists(filepath.Join(h.Dir, file)) {
						value = tengo.TrueValue
					}

					return value, nil
				},
			},
		},
	}
}
