package engine

import (
	ffuf "github.com/analog-substance/ffufwrap"
	"github.com/analog-substance/tengo/v2"
)

func (s *Script) FfufModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"fuzzer": &tengo.UserFunction{
			Name:  "fuzzer",
			Value: s.ffufFuzzer,
		},
		"default_strategy": &tengo.String{
			Value: string(ffuf.DefaultStrategy),
		},
		"greedy_strategy": &tengo.String{
			Value: string(ffuf.GreedyStrategy),
		},
		"basic_strategy": &tengo.String{
			Value: string(ffuf.BasicStrategy),
		},
		"advanced_strategy": &tengo.String{
			Value: string(ffuf.AdvancedStrategy),
		},
		"or_operator": &tengo.String{
			Value: string(ffuf.OrOperator),
		},
		"and_operator": &tengo.String{
			Value: string(ffuf.AndOperator),
		},
		"cluster_bomb": &tengo.String{
			Value: string(ffuf.ModeClusterBomb),
		},
		"pitch_fork": &tengo.String{
			Value: string(ffuf.ModePitchFork),
		},
		"sniper": &tengo.String{
			Value: string(ffuf.ModeSniper),
		},
		"format_all": &tengo.String{
			Value: string(ffuf.FormatAll),
		},
		"format_json": &tengo.String{
			Value: string(ffuf.FormatJSON),
		},
		"format_ejson": &tengo.String{
			Value: string(ffuf.FormatEJSON),
		},
		"format_html": &tengo.String{
			Value: string(ffuf.FormatHTML),
		},
		"format_md": &tengo.String{
			Value: string(ffuf.FormatMarkdown),
		},
		"format_csv": &tengo.String{
			Value: string(ffuf.FormatCSV),
		},
		"format_ecsv": &tengo.String{
			Value: string(ffuf.FormatECSV),
		},
	}
}

func (s *Script) ffufFuzzer(args ...tengo.Object) (tengo.Object, error) {
	return newFfufFuzzer(s), nil
}
