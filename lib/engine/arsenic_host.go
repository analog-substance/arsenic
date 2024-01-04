package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengomod/interop"
	"github.com/analog-substance/tengomod/types"
)

type ArsenicHost struct {
	types.PropObject
	Value *host.Host
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

func (h *ArsenicHost) urls(args interop.ArgMap) (tengo.Object, error) {
	protocols, _ := args.GetStringSlice("protocols")
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

	return interop.GoStrSliceToTArray(urls), nil
}

func (h *ArsenicHost) fileExists(args interop.ArgMap) (tengo.Object, error) {
	file, _ := args.GetString("path")

	exists := fileutil.FileExists(filepath.Join(h.Value.Dir, file))
	return interop.GoBoolToTBool(exists), nil
}

func (h *ArsenicHost) contentDiscoveryURLs(args interop.ArgMap) (tengo.Object, error) {
	patterns, _ := args.GetStringSlice("patterns")
	codes, _ := args.GetIntSlice("codes")

	var files []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return interop.GoErrToTErr(err), nil
		}
		files = append(files, matches...)
	}

	allResults, err := gocdp.SmartParseFiles(files)
	if err != nil {
		return interop.GoErrToTErr(err), nil
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

	return interop.GoStrSliceToTArray(urls), nil
}

func (h *ArsenicHost) sync(args ...tengo.Object) (tengo.Object, error) {
	err := h.Value.SyncMetadata(host.SyncOptions{
		IPAddresses: true,
		Hostnames:   true,
		Ports:       true,
		Flags:       true,
	})
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	h.Value.SaveMetadata()
	return nil, nil
}

func (h *ArsenicHost) tcpPorts() tengo.Object {
	var ports []tengo.Object
	for _, port := range h.Value.Metadata.Ports {
		if port.Protocol == "tcp" {
			ports = append(ports, makeArsenicPort(port))
		}
	}

	return &tengo.ImmutableArray{
		Value: ports,
	}
}

func (h *ArsenicHost) udpPorts() tengo.Object {
	var ports []tengo.Object
	for _, port := range h.Value.Metadata.Ports {
		if port.Protocol == "udp" {
			ports = append(ports, makeArsenicPort(port))
		}
	}

	return &tengo.ImmutableArray{
		Value: ports,
	}
}

func makeArsenicHost(h *host.Host) *ArsenicHost {
	arsenicHost := &ArsenicHost{
		Value: h,
	}

	objectMap := map[string]tengo.Object{
		"has_flags": &tengo.UserFunction{
			Name:  "has_flags",
			Value: interop.FuncASvRB(h.Metadata.HasFlags),
		},
		"has_any_port": &tengo.UserFunction{
			Name:  "has_any_port",
			Value: stdlib.FuncARB(h.Metadata.HasAnyPort),
		},
		"files": &tengo.UserFunction{
			Name:  "files",
			Value: interop.FuncASvRSsE(h.Files),
		},
		"urls": &interop.AdvFunction{
			Name:  "urls",
			Args:  []interop.AdvArg{interop.StrSliceArg("protocols", true)},
			Value: arsenicHost.urls,
		},
		"file_exists": &interop.AdvFunction{
			Name:    "file_exists",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("path")},
			Value:   arsenicHost.fileExists,
		},
		"content_discovery_urls": &interop.AdvFunction{
			Name:    "content_discovery_urls",
			NumArgs: interop.ExactArgs(2),
			Args:    []interop.AdvArg{interop.StrSliceArg("patterns", false), interop.IntSliceArg("codes", false)},
			Value:   arsenicHost.contentDiscoveryURLs,
		},
		"sync": &tengo.UserFunction{
			Name:  "sync",
			Value: arsenicHost.sync,
		},
	}

	properties := map[string]types.Property{
		"dir": {
			Get: func() tengo.Object {
				return &tengo.String{
					Value: h.Dir,
				}
			},
		},
		"name": {
			Get: func() tengo.Object {
				return &tengo.String{
					Value: h.Metadata.Name,
				}
			},
		},
		"ip_addresses": {
			Get: func() tengo.Object {
				return interop.GoStrSliceToTArray(h.Metadata.IPAddresses)
			},
		},
		"hostnames": {
			Get: func() tengo.Object {
				return interop.GoStrSliceToTArray(h.Metadata.Hostnames)
			},
		},
		"tcp_ports": {
			Get: arsenicHost.tcpPorts,
		},
		"udp_ports": {
			Get: arsenicHost.udpPorts,
		},
	}

	arsenicHost.PropObject = types.PropObject{
		ObjectMap:  objectMap,
		Properties: properties,
	}

	return arsenicHost
}

type ArsenicPort struct {
	types.PropObject
	Value host.Port
}

func (p *ArsenicPort) TypeName() string {
	return "arsenic-port"
}

// String should return a string representation of the type's value.
func (p *ArsenicPort) String() string {
	return fmt.Sprintf("%d/%s %s", p.Value.ID, p.Value.Protocol, p.Value.Service)
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (p *ArsenicPort) IsFalsy() bool {
	return p.Value.ID == 0
}

// CanIterate should return whether the Object can be Iterated.
func (p *ArsenicPort) CanIterate() bool {
	return false
}

func makeArsenicPort(port host.Port) *ArsenicPort {
	p := &ArsenicPort{
		Value: port,
	}

	properties := map[string]types.Property{
		"port":     types.StaticProperty(&tengo.Int{Value: int64(port.ID)}),
		"protocol": types.StaticProperty(interop.GoStrToTStr(port.Protocol)),
		"service":  types.StaticProperty(interop.GoStrToTStr(port.Service)),
	}

	p.PropObject = types.PropObject{
		ObjectMap:  make(map[string]tengo.Object),
		Properties: properties,
	}

	return p
}
