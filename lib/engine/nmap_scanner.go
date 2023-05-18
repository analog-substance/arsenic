package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v3"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
)

// NmapScanner is the tengo wrapper object for nmap.Scanner
type NmapScanner struct {
	tengo.ObjectImpl
	Value     *nmap.Scanner
	objectMap map[string]tengo.Object
	script    *Script
}

// addOptionA transform a function of 'func() nmap.Option' signature
// into tengo CallableFunc type.
func (s *NmapScanner) addOptionA(fn func() nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		option := fn()

		s.Value.AddOptions(option)
		return s, nil
	}
}

// addOptionAS transform a function of 'func(string) nmap.Option' signature
// into tengo CallableFunc type.
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
		return s, nil
	}
}

// addOptionAI transform a function of 'func(int) nmap.Option' signature
// into tengo CallableFunc type.
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
		return s, nil
	}
}

// addOptionAI16 transform a function of 'func(int16) nmap.Option' signature
// into tengo CallableFunc type.
func (s *NmapScanner) addOptionAI16(fn func(int16) nmap.Option) tengo.CallableFunc {
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

		option := fn(int16(i1))

		s.Value.AddOptions(option)
		return s, nil
	}
}

// addOptionAD transform a function of 'func(time.Duration) nmap.Option' signature
// into tengo CallableFunc type.
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
		return s, nil
	}
}

// addOptionASv transform a function of 'func(...string) nmap.Option' signature
// into tengo CallableFunc type.
func (s *NmapScanner) addOptionASv(fn func(...string) nmap.Option) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) == 0 {
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
		return s, nil
	}
}

// aliasFunc is used to call the same tengo function using a different name
func (s *NmapScanner) aliasFunc(name string, src string) *tengo.UserFunction {
	return aliasFunc(s, name, src)
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
	scanner, err := nmap.NewScanner(context.Background())
	if err != nil {
		return nil, err
	}

	scanner.Streamer(os.Stdout)

	nmapScanner := &NmapScanner{
		Value:  scanner,
		script: s,
	}

	objectMap := map[string]tengo.Object{
		"disabled_dns_resolution": &tengo.UserFunction{
			Name:  "disabled_dns_resolution",
			Value: nmapScanner.addOptionA(nmap.WithDisabledDNSResolution),
		},
		"n": nmapScanner.aliasFunc("n", "disabled_dns_resolution"),
		"list_scan": &tengo.UserFunction{
			Name:  "list_scan",
			Value: nmapScanner.addOptionA(nmap.WithListScan),
		},
		"sL": nmapScanner.aliasFunc("sL", "list_scan"),
		"open_only": &tengo.UserFunction{
			Name:  "open_only",
			Value: nmapScanner.addOptionA(nmap.WithOpenOnly),
		},
		"open": nmapScanner.aliasFunc("open", "open_only"),
		"ping_scan": &tengo.UserFunction{
			Name:  "ping_scan",
			Value: nmapScanner.addOptionA(nmap.WithPingScan),
		},
		"sn": nmapScanner.aliasFunc("sn", "ping_scan"),
		"service_info": &tengo.UserFunction{
			Name:  "service_info",
			Value: nmapScanner.addOptionA(nmap.WithServiceInfo),
		},
		"sV": nmapScanner.aliasFunc("sV", "service_info"),
		"skip_host_discovery": &tengo.UserFunction{
			Name:  "skip_host_discovery",
			Value: nmapScanner.addOptionA(nmap.WithSkipHostDiscovery),
		},
		"Pn": nmapScanner.aliasFunc("Pn", "skip_host_discovery"),
		"system_dns": &tengo.UserFunction{
			Name:  "system_dns",
			Value: nmapScanner.addOptionA(nmap.WithSystemDNS),
		},
		"udp_scan": &tengo.UserFunction{
			Name:  "udp_scan",
			Value: nmapScanner.addOptionA(nmap.WithUDPScan),
		},
		"sU": nmapScanner.aliasFunc("sU", "udp_scan"),
		"version_intensity": &tengo.UserFunction{
			Name:  "version_intensity",
			Value: nmapScanner.addOptionAI16(nmap.WithVersionIntensity),
		},
		"grep_output": &tengo.UserFunction{
			Name:  "grep_output",
			Value: nmapScanner.addOptionAS(nmap.WithGrepOutput),
		},
		"oG": nmapScanner.aliasFunc("oG", "grep_output"),
		"nmap_output": &tengo.UserFunction{
			Name:  "nmap_output",
			Value: nmapScanner.addOptionAS(nmap.WithNmapOutput),
		},
		"oN": nmapScanner.aliasFunc("oN", "nmap_output"),
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

				nmapScanner.Value.ToFile(s1)

				return nmapScanner, nil
			},
		},
		"oX": nmapScanner.aliasFunc("oX", "xml_output"),
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
				nmapScanner.Value.ToFile(fmt.Sprintf("%s.xml", s1))

				return nmapScanner, nil
			},
		},
		"oA": nmapScanner.aliasFunc("oA", "all_output"),
		"stylesheet": &tengo.UserFunction{
			Name:  "stylesheet",
			Value: nmapScanner.addOptionAS(nmap.WithStylesheet),
		},
		"target_input": &tengo.UserFunction{
			Name:  "target_input",
			Value: nmapScanner.addOptionAS(nmap.WithTargetInput),
		},
		"iL": nmapScanner.aliasFunc("iL", "target_input"),
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
		"min_rate": &tengo.UserFunction{
			Name:  "min_rate",
			Value: nmapScanner.addOptionAI(nmap.WithMinRate),
		},
		"top_ports": &tengo.UserFunction{
			Name:  "top_ports",
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
				return nmapScanner, nil
			},
		},
		"T0": &tengo.UserFunction{
			Name: "T0",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingSlowest)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"T1": &tengo.UserFunction{
			Name: "T1",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingSneaky)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"T2": &tengo.UserFunction{
			Name: "T2",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingPolite)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"T3": &tengo.UserFunction{
			Name: "T3",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingNormal)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"T4": &tengo.UserFunction{
			Name: "T4",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingAggressive)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"T5": &tengo.UserFunction{
			Name: "T5",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				option := nmap.WithTimingTemplate(nmap.TimingFastest)
				nmapScanner.Value.AddOptions(option)
				return nmapScanner, nil
			},
		},
		"aggressive_scan": &tengo.UserFunction{
			Name:  "aggressive_scan",
			Value: nmapScanner.addOptionA(nmap.WithAggressiveScan),
		},
		"A": nmapScanner.aliasFunc("A", "aggressive_scan"),
		"args": &tengo.UserFunction{
			Name:  "args",
			Value: stdlib.FuncARSs(nmapScanner.Value.Args),
		},
		"custom_args": &tengo.UserFunction{
			Name:  "custom_args",
			Value: nmapScanner.addOptionASv(nmap.WithCustomArguments),
		},
		"privileged": &tengo.UserFunction{
			Name:  "privileged",
			Value: nmapScanner.addOptionA(nmap.WithPrivileged),
		},
		"run": &tengo.UserFunction{
			Name: "run",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				run, warnings, err := nmapScanner.Value.Run()
				if err != nil {
					return toError(fmt.Errorf("%v: %s", err, strings.Join(*warnings, "\n"))), nil
				}

				return makeNmapRun(run), nil
			},
		},
	}

	nmapScanner.objectMap = objectMap
	return nmapScanner, nil
}

// NmapRun represents a simple tengo object wrapper for *nmap.Run
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
		"ports": sliceToIntArray(ports),
	}

	nmapRun.objectMap = objectMap
	return nmapRun
}
