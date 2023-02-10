package engine

import (
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) NmapModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"scanner":           &tengo.UserFunction{Name: "scanner", Value: s.nmapScanner},
		"timing_slowest":    &tengo.Int{Value: 0},
		"timing_sneaky":     &tengo.Int{Value: 1},
		"timing_polite":     &tengo.Int{Value: 2},
		"timing_normal":     &tengo.Int{Value: 3},
		"timing_aggressive": &tengo.Int{Value: 4},
		"timing_fastest":    &tengo.Int{Value: 5},
	}
}

func (s *Script) nmapScanner(args ...tengo.Object) (tengo.Object, error) {
	scanner, err := makeNmapScanner(s)
	if err != nil {
		return nil, err
	}

	return scanner, nil
}
