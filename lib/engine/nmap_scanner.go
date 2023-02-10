package engine

import (
	"fmt"
	"time"

	"github.com/Ullaakut/nmap/v2"
	"github.com/analog-substance/tengo/v2"
)

type NmapScanner struct {
	tengo.ObjectImpl
	Value     *nmap.Scanner
	objectMap map[string]tengo.Object
	script    *Script
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
	}

	nmapScanner.objectMap = objectMap
	return nmapScanner, nil
}
