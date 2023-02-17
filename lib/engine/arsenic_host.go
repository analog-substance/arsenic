package engine

import (
	"path/filepath"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

type ArsenicHost struct {
	tengo.ObjectImpl
	Value     *host.Host
	objectMap map[string]tengo.Object
}

func (h *ArsenicHost) TypeName() string {
	return "arsenic-host"
}

// String should return a string representation of the type's value.
func (h *ArsenicHost) String() string {
	return h.Value.Metadata.Name
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (h *ArsenicHost) IsFalsy() bool {
	return h.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (h *ArsenicHost) CanIterate() bool {
	return false
}

func (h *ArsenicHost) IndexGet(index tengo.Object) (tengo.Object, error) {
	strIdx, ok := tengo.ToString(index)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}

	res, ok := h.objectMap[strIdx]
	if !ok {
		res = tengo.UndefinedValue
	}
	return res, nil
}

func (h *ArsenicHost) urls(args ...tengo.Object) (tengo.Object, error) {
	protocols, err := sliceToStringSlice(args)
	if err != nil {
		return nil, err
	}

	if len(protocols) == 0 {
		protocols = append(protocols, "all")
	}

	var urls []string
	for _, hostURL := range h.Value.URLs() {
		for _, proto := range protocols {
			if strings.HasPrefix(hostURL, proto) || proto == "all" {
				urls = append(urls, hostURL)
			}
		}
	}

	return sliceToStringArray(urls), nil
}

func (h *ArsenicHost) fileExists(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	file, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	value := tengo.FalseValue
	if util.FileExists(filepath.Join(h.Value.Dir, file)) {
		value = tengo.TrueValue
	}

	return value, nil
}

func (h *ArsenicHost) contentDiscoveryURLs(args ...tengo.Object) (tengo.Object, error) {
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

func makeArsenicHost(h *host.Host) *ArsenicHost {
	arsenicHost := &ArsenicHost{
		Value: h,
	}

	objectMap := map[string]tengo.Object{
		"dir": &tengo.String{
			Value: h.Dir,
		},
		"name": &tengo.String{
			Value: h.Metadata.Name,
		},
		"has_flags": &tengo.UserFunction{
			Name:  "has_flags",
			Value: funcASvRB(h.Metadata.HasFlags),
		},
		"has_any_port": &tengo.UserFunction{
			Name:  "has_any_port",
			Value: stdlib.FuncARB(h.Metadata.HasAnyPort),
		},
		"files": &tengo.UserFunction{
			Name:  "files",
			Value: funcASvRSsE(h.Files),
		},
		"urls": &tengo.UserFunction{
			Name:  "urls",
			Value: arsenicHost.urls,
		},
		"file_exists": &tengo.UserFunction{
			Name:  "file_exists",
			Value: arsenicHost.fileExists,
		},
		"content_discovery_urls": &tengo.UserFunction{
			Name:  "content_discovery_urls",
			Value: arsenicHost.contentDiscoveryURLs,
		},
	}

	arsenicHost.objectMap = objectMap

	return arsenicHost
}
