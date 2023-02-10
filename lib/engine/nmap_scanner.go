package engine

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

type NmapScanner struct {
	tengo.ObjectImpl
	Value         *nmap.Scanner
	objectMap     map[string]tengo.Object
	script        *Script
	xmlOutputName string
}

func (s *NmapScanner) addOptionA(fn func() nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		option := fn()

		s.Value.AddOptions(option)
		return nil, nil
	}
}

func (s *NmapScanner) addOptionAS(fn func(string) nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		option := fn(s1)

		s.Value.AddOptions(option)
		return nil, nil
	}
}

func (s *NmapScanner) addOptionAI(fn func(int) nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		i1, ok := tengo.ToInt(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "int(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		option := fn(i1)

		s.Value.AddOptions(option)
		return nil, nil
	}
}

func (s *NmapScanner) addOptionAD(fn func(time.Duration) nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}

		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "first",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}

		dur, err := time.ParseDuration(s1)
		if err != nil {
			return nil, err
		}

		option := fn(dur)

		s.Value.AddOptions(option)
		return nil, nil
	}
}

func (s *NmapScanner) addOptionASv(fn func(...string) nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		var strings []string
		for i, arg := range args {
			str, ok := tengo.ToString(arg)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     fmt.Sprintf("#%d arg", i),
					Expected: "string(compatible)",
					Found:    arg.TypeName(),
				}
			}

			strings = append(strings, str)
		}

		option := fn(strings...)

		s.Value.AddOptions(option)
		return nil, nil
	}
}

// TypeName should return the name of the type.
func (s *NmapScanner) TypeName() string {
	return "nmap-scanner"
}

// String should return a string representation of the type's value.
func (s *NmapScanner) String() string {
	return strings.Join(s.Value.Args(), " ")
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (s *NmapScanner) IsFalsy() bool {
	return s.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (s *NmapScanner) CanIterate() bool {
	return false
}

func (s *NmapScanner) IndexGet(index tengo.Object) (value tengo.Object, err error) {
	strIdx, ok := tengo.ToString(index)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}

	res, ok := s.objectMap[strIdx]
	if !ok {
		res = tengo.UndefinedValue
	}
	return res, nil
}

func makeNmapScanner(s *Script) (*NmapScanner, error) {
	scanner, err := nmap.NewScanner()
	if err != nil {
		return nil, err
	}

	nmapScanner := &NmapScanner{
		Value:  scanner,
		script: s,
	}

	objectMap := map[string]tengo.Object{
		"disabled_dns_resolution": &tengo.UserFunction{
			Name:  "disabled_dns_resolution",
			Value: nmapScanner.addOptionA(nmap.WithDisabledDNSResolution),
		},
		"list_scan": &tengo.UserFunction{
			Name:  "list_scan",
			Value: nmapScanner.addOptionA(nmap.WithListScan),
		},
		"open_only": &tengo.UserFunction{
			Name:  "open_only",
			Value: nmapScanner.addOptionA(nmap.WithOpenOnly),
		},
		"ping_scan": &tengo.UserFunction{
			Name:  "ping_scan",
			Value: nmapScanner.addOptionA(nmap.WithPingScan),
		},
		"service_info": &tengo.UserFunction{
			Name:  "service_info",
			Value: nmapScanner.addOptionA(nmap.WithServiceInfo),
		},
		"skip_host_discovery": &tengo.UserFunction{
			Name:  "skip_host_discovery",
			Value: nmapScanner.addOptionA(nmap.WithSkipHostDiscovery),
		},
		"system_dns": &tengo.UserFunction{
			Name:  "system_dns",
			Value: nmapScanner.addOptionA(nmap.WithSystemDNS),
		},
		"udp_scan": &tengo.UserFunction{
			Name:  "udp_scan",
			Value: nmapScanner.addOptionA(nmap.WithUDPScan),
		},
		"grep_output": &tengo.UserFunction{
			Name:  "grep_output",
			Value: nmapScanner.addOptionAS(nmap.WithGrepOutput),
		},
		"nmap_output": &tengo.UserFunction{
			Name:  "nmap_output",
			Value: nmapScanner.addOptionAS(nmap.WithNmapOutput),
		},
		"xml_output": &tengo.UserFunction{
			Name: "xml_output",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				s1, ok := tengo.ToString(args[0])
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     "first",
						Expected: "string(compatible)",
						Found:    args[0].TypeName(),
					}
				}

				nmapScanner.xmlOutputName = s1

				return nil, nil
			},
		},
		"all_output": &tengo.UserFunction{
			Name: "all_output",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				s1, ok := tengo.ToString(args[0])
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     "first",
						Expected: "string(compatible)",
						Found:    args[0].TypeName(),
					}
				}

				nmapScanner.Value.AddOptions(
					nmap.WithGrepOutput(fmt.Sprintf("%s.gnmap", s1)),
					nmap.WithNmapOutput(fmt.Sprintf("%s.nmap", s1)),
				)
				nmapScanner.xmlOutputName = fmt.Sprintf("%s.xml", s1)

				return nil, nil
			},
		},
		"stylesheet": &tengo.UserFunction{
			Name:  "stylesheet",
			Value: nmapScanner.addOptionAS(nmap.WithStylesheet),
		},
		"target_input": &tengo.UserFunction{
			Name:  "target_input",
			Value: nmapScanner.addOptionAS(nmap.WithTargetInput),
		},
		"host_timeout": &tengo.UserFunction{
			Name:  "host_timeout",
			Value: nmapScanner.addOptionAD(nmap.WithHostTimeout),
		},
		"max_rtt_timeout": &tengo.UserFunction{
			Name:  "max_rtt_timeout",
			Value: nmapScanner.addOptionAD(nmap.WithMaxRTTTimeout),
		},
		"max_rate": &tengo.UserFunction{
			Name:  "max_rate",
			Value: nmapScanner.addOptionAI(nmap.WithMaxRate),
		},
		"most_common_ports": &tengo.UserFunction{
			Name:  "most_common_ports",
			Value: nmapScanner.addOptionAI(nmap.WithMostCommonPorts),
		},
		"ports": &tengo.UserFunction{
			Name:  "ports",
			Value: nmapScanner.addOptionASv(nmap.WithPorts),
		},
		"targets": &tengo.UserFunction{
			Name:  "targets",
			Value: nmapScanner.addOptionASv(nmap.WithTargets),
		},
		"timing_template": &tengo.UserFunction{
			Name: "timing_template",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				i1, ok := tengo.ToInt(args[0])
				if !ok {
					return nil, tengo.ErrInvalidArgumentType{
						Name:     "first",
						Expected: "int(compatible)",
						Found:    args[0].TypeName(),
					}
				}

				option := nmap.WithTimingTemplate(nmap.Timing(i1))
				nmapScanner.Value.AddOptions(option)
				return nil, nil
			},
		},
		"args": &tengo.UserFunction{
			Name:  "args",
			Value: stdlib.FuncARSs(nmapScanner.Value.Args),
		},
		"run": &tengo.UserFunction{
			Name: "run",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				run, warnings, err := nmapScanner.Value.Run()
				if err != nil {
					return toError(fmt.Errorf("%v: %s", err, strings.Join(warnings, "\n"))), nil
				}

				if nmapScanner.xmlOutputName != "" {
					err = run.ToFile(nmapScanner.xmlOutputName)
					if err != nil {
						return toError(err), nil
					}
				}
				return makeNmapRun(run), nil
			},
		},
	}

	nmapScanner.objectMap = objectMap
	return nmapScanner, nil
}

type NmapRun struct {
	tengo.ObjectImpl
	Value     *nmap.Run
	objectMap map[string]tengo.Object
}

// TypeName should return the name of the type.
func (s *NmapRun) TypeName() string {
	return "nmap-run"
}

// String should return a string representation of the type's value.
func (r *NmapRun) String() string {
	bytes, _ := io.ReadAll(r.Value.ToReader())
	return string(bytes)
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (r *NmapRun) IsFalsy() bool {
	return r.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (r *NmapRun) CanIterate() bool {
	return false
}

func (r *NmapRun) IndexGet(index tengo.Object) (value tengo.Object, err error) {
	strIdx, ok := tengo.ToString(index)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}

	res, ok := r.objectMap[strIdx]
	if !ok {
		res = tengo.UndefinedValue
	}
	return res, nil
}

func makeNmapRun(run *nmap.Run) *NmapRun {
	nmapRun := &NmapRun{
		Value: run,
	}

	var ports []int
	for _, h := range nmapRun.Value.Hosts {
		for _, p := range h.Ports {
			ports = append(ports, int(p.ID))
		}
	}

	// Currently only need ports, probably will want to implement more
	objectMap := map[string]tengo.Object{
		"ports": toIntArray(ports),
	}

	nmapRun.objectMap = objectMap
	return nmapRun
}
