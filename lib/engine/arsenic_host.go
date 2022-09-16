package engine

import (
	"path/filepath"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
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
			"content_discovery_urls": &tengo.UserFunction{
				Name: "content_discovery_urls",
				Value: func(args ...tengo.Object) (tengo.Object, error) {
					if len(args) != 2 {
						return toError(tengo.ErrWrongNumArguments), nil
					}

					patternsArray, ok := args[0].(*tengo.Array)
					if !ok {
						return toError(tengo.ErrInvalidArgumentType{
							Name:     "patterns",
							Expected: "string",
							Found:    args[0].TypeName(),
						}), nil
					}

					patterns, err := toStringSlice(patternsArray)
					if err != nil {
						return toError(err), nil
					}

					codesArray, ok := args[1].(*tengo.Array)
					if !ok {
						return toError(tengo.ErrInvalidArgumentType{
							Name:     "codes",
							Expected: "string",
							Found:    args[1].TypeName(),
						}), nil
					}

					codes, err := toIntSlice(codesArray)
					if err != nil {
						return toError(err), nil
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

					return toStringArray(urls), nil
				},
			},
		},
	}
}
