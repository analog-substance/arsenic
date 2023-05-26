package engine

import (
	"strings"

	ffuf "github.com/analog-substance/ffufwrap"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengo/v2/stdlib"
	"github.com/analog-substance/tengo/v2/stdlib/json"
	"github.com/analog-substance/tengomod/interop"
	"github.com/analog-substance/tengomod/types"
)

type Fuzzer struct {
	types.PropObject
	Value  *ffuf.Fuzzer
	script *Script
}

func (f *Fuzzer) TypeName() string {
	return "ffuf-fuzzer"
}

// String should return a string representation of the type's value.
func (f *Fuzzer) String() string {
	return strings.Join(f.Value.Args(), " ")
}

// IsFalsy should return true if the value of the type should be considered
// as falsy.
func (f *Fuzzer) IsFalsy() bool {
	return f.Value == nil
}

// CanIterate should return whether the Object can be Iterated.
func (f *Fuzzer) CanIterate() bool {
	return false
}

func (f *Fuzzer) funcASRF(fn func(string) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.ExactArgs(1),
		Args:    []interop.AdvArg{interop.StrArg("first")},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			s1, _ := args.GetString("first")

			fn(s1)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) funcAIRF(fn func(int) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.ExactArgs(1),
		Args:    []interop.AdvArg{interop.IntArg("first")},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			i1, _ := args.GetInt("first")

			fn(i1)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) funcASSRF(fn func(string, string) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.ExactArgs(2),
		Args:    []interop.AdvArg{interop.StrArg("first"), interop.StrArg("second")},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			s1, _ := args.GetString("first")
			s2, _ := args.GetString("second")

			fn(s1, s2)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) funcASvRF(fn func(...string) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.MinArgs(1),
		Args:    []interop.AdvArg{interop.StrSliceArg("first", true)},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			slice, _ := args.GetStringSlice("first")

			fn(slice...)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) funcASsRF(fn func([]string) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.ExactArgs(1),
		Args:    []interop.AdvArg{interop.StrSliceArg("first", false)},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			slice, _ := args.GetStringSlice("first")

			fn(slice)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) funcARF(fn func() *ffuf.Fuzzer) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		fn()
		return f, nil
	}
}

func (f *Fuzzer) funcASMSRF(fn func(map[string]string) *ffuf.Fuzzer) tengo.CallableFunc {
	advFunc := interop.AdvFunction{
		NumArgs: interop.ExactArgs(1),
		Args:    []interop.AdvArg{interop.StrMapStrArg("first")},
		Value: func(args interop.ArgMap) (tengo.Object, error) {
			m, _ := args.GetStrMapStr("first")

			fn(m)
			return f, nil
		},
	}
	return advFunc.Call
}

func (f *Fuzzer) recursionStrategy(args interop.ArgMap) (tengo.Object, error) {
	strategy, _ := args.GetString("strategy")

	f.Value.RecursionStrategy(ffuf.RecursionStrategy(strategy))
	return f, nil
}

func (f *Fuzzer) autoCalibrateStrategy(args interop.ArgMap) (tengo.Object, error) {
	strategy, _ := args.GetString("strategy")

	f.Value.AutoCalibrateStrategy(ffuf.AutoCalibrateStrategy(strategy))
	return f, nil
}

func (f *Fuzzer) matchOperator(args interop.ArgMap) (tengo.Object, error) {
	operator, _ := args.GetString("operator")

	f.Value.MatchOperator(ffuf.SetOperator(operator))
	return f, nil
}

func (f *Fuzzer) filterOperator(args interop.ArgMap) (tengo.Object, error) {
	operator, _ := args.GetString("operator")

	f.Value.FilterOperator(ffuf.SetOperator(operator))
	return f, nil
}

func (f *Fuzzer) postJSON(args interop.ArgMap) (tengo.Object, error) {
	body, _ := args.GetObject("body")
	bytes, err := json.Encode(body)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	f.Value.PostString(string(bytes))
	return f, nil
}

func (f *Fuzzer) wordlistMode(args interop.ArgMap) (tengo.Object, error) {
	mode, _ := args.GetString("mode")

	f.Value.WordlistMode(ffuf.WordlistMode(mode))
	return f, nil
}

func (f *Fuzzer) outputFormat(args interop.ArgMap) (tengo.Object, error) {
	format, _ := args.GetString("format")

	f.Value.OutputFormat(ffuf.OutputFormat(format))
	return f, nil
}

func (f *Fuzzer) customArguments(args interop.ArgMap) (tengo.Object, error) {
	slice, _ := args.GetStringSlice("args")

	f.Value.CustomArguments(slice...)
	return f, nil
}

func (f *Fuzzer) clone(args ...tengo.Object) (tengo.Object, error) {
	return makeFfufFuzzer(f.script, f.Value.Clone(f.script.ctx)), nil
}

func (f *Fuzzer) aliasFunc(name string, src string) *tengo.UserFunction {
	return aliasFunc(f, name, src)
}

func makeFfufFuzzer(s *Script, f *ffuf.Fuzzer) *Fuzzer {
	fuzzer := &Fuzzer{
		Value:  f,
		script: s,
	}

	objectMap := map[string]tengo.Object{
		"clone": &tengo.UserFunction{
			Name:  "clone",
			Value: fuzzer.clone,
		},
		"binary_path": &tengo.UserFunction{
			Name:  "binary_path",
			Value: fuzzer.funcASRF(f.BinaryPath),
		},
		"auto_append_keyword": &tengo.UserFunction{
			Name:  "auto_append_keyword",
			Value: fuzzer.funcARF(f.AutoAppendKeyword),
		},
		"headers": &tengo.UserFunction{
			Name:  "headers",
			Value: fuzzer.funcASMSRF(f.Headers),
		},
		"header": &tengo.UserFunction{
			Name:  "header",
			Value: fuzzer.funcASSRF(f.Header),
		},
		"H": fuzzer.aliasFunc("H", "header"),
		"recursion_depth": &tengo.UserFunction{
			Name:  "recursion_depth",
			Value: fuzzer.funcAIRF(f.RecursionDepth),
		},
		"recursion": &tengo.UserFunction{
			Name:  "recursion",
			Value: fuzzer.funcARF(f.Recursion),
		},
		"recursion_strategy": &interop.AdvFunction{
			Name:    "recursion_strategy",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("strategy")},
			Value:   fuzzer.recursionStrategy,
		},
		"replay_proxy": &tengo.UserFunction{
			Name:  "replay_proxy",
			Value: fuzzer.funcARF(f.ReplayProxy),
		},
		"sni": &tengo.UserFunction{
			Name:  "sni",
			Value: fuzzer.funcARF(f.SNI),
		},
		"timeout": &tengo.UserFunction{
			Name:  "timeout",
			Value: fuzzer.funcAIRF(f.Timeout),
		},
		"auto_calibrate": &tengo.UserFunction{
			Name:  "auto_calibrate",
			Value: fuzzer.funcARF(f.AutoCalibrate),
		},
		"ac": fuzzer.aliasFunc("ac", "auto_calibrate"),
		"custom_auto_calibrate": &tengo.UserFunction{
			Name:  "custom_auto_calibrate",
			Value: fuzzer.funcASvRF(f.CustomAutoCalibrate),
		},
		"acc": fuzzer.aliasFunc("acc", "custom_auto_calibrate"),
		"per_host_auto_calibrate": &tengo.UserFunction{
			Name:  "per_host_auto_calibrate",
			Value: fuzzer.funcARF(f.PerHostAutoCalibrate),
		},
		"ach": fuzzer.aliasFunc("ach", "per_host_auto_calibrate"),
		"auto_calibrate_strategy": &interop.AdvFunction{
			Name:  "auto_calibrate_strategy",
			Value: fuzzer.autoCalibrateStrategy,
		},
		"acs": fuzzer.aliasFunc("acs", "auto_calibrate_strategy"),
		"colorize_output": &tengo.UserFunction{
			Name:  "colorize_output",
			Value: fuzzer.funcARF(f.ColorizeOutput),
		},
		"c": fuzzer.aliasFunc("c", "colorize_output"),
		"config_file": &tengo.UserFunction{
			Name:  "config_file",
			Value: fuzzer.funcASRF(f.ConfigFile),
		},
		"print_json": &tengo.UserFunction{
			Name:  "print_json",
			Value: fuzzer.funcARF(f.PrintJSON),
		},
		"max_total_time": &tengo.UserFunction{
			Name:  "max_total_time",
			Value: fuzzer.funcAIRF(f.MaxTotalTime),
		},
		"max_job_time": &tengo.UserFunction{
			Name:  "max_job_time",
			Value: fuzzer.funcAIRF(f.MaxJobTime),
		},
		"non_interactive": &tengo.UserFunction{
			Name:  "non_interactive",
			Value: fuzzer.funcARF(f.NonInteractive),
		},
		"request_rate": &tengo.UserFunction{
			Name:  "request_rate",
			Value: fuzzer.funcAIRF(f.RequestRate),
		},
		"silent": &tengo.UserFunction{
			Name:  "silent",
			Value: fuzzer.funcARF(f.Silent),
		},
		"stop_on_all_errors": &tengo.UserFunction{
			Name:  "stop_on_all_errors",
			Value: fuzzer.funcARF(f.StopOnAllErrors),
		},
		"sa": fuzzer.aliasFunc("sa", "stop_on_all_errors"),
		"stop_on_spurious_errors": &tengo.UserFunction{
			Name:  "stop_on_spurious_errors",
			Value: fuzzer.funcARF(f.StopOnSpuriousErrors),
		},
		"se": fuzzer.aliasFunc("se", "stop_on_spurious_errors"),
		"stop_on_forbidden": &tengo.UserFunction{
			Name:  "stop_on_forbidden",
			Value: fuzzer.funcARF(f.StopOnForbidden),
		},
		"sf": fuzzer.aliasFunc("sf", "stop_on_forbidden"),
		"threads": &tengo.UserFunction{
			Name:  "threads",
			Value: fuzzer.funcAIRF(f.Threads),
		},
		"verbose": &tengo.UserFunction{
			Name:  "verbose",
			Value: fuzzer.funcARF(f.Verbose),
		},
		"method": &tengo.UserFunction{
			Name:  "method",
			Value: fuzzer.funcASRF(f.Method),
		},
		"delay": &tengo.UserFunction{
			Name:  "delay",
			Value: fuzzer.funcAIRF(f.Delay),
		},
		"exts": &tengo.UserFunction{
			Name:  "exts",
			Value: fuzzer.funcASsRF(f.Exts),
		},
		"match_codes": &tengo.UserFunction{
			Name:  "match_codes",
			Value: fuzzer.funcASvRF(f.MatchCodes),
		},
		"match_lines": &tengo.UserFunction{
			Name:  "match_lines",
			Value: fuzzer.funcAIRF(f.MatchLines),
		},
		"match_size": &tengo.UserFunction{
			Name:  "match_size",
			Value: fuzzer.funcAIRF(f.MatchSize),
		},
		"match_words": &tengo.UserFunction{
			Name:  "match_words",
			Value: fuzzer.funcAIRF(f.MatchWords),
		},
		"match_regex": &tengo.UserFunction{
			Name:  "match_regex",
			Value: fuzzer.funcASRF(f.MatchRegex),
		},
		"match_time": &tengo.UserFunction{
			Name:  "match_time",
			Value: fuzzer.funcAIRF(f.MatchTime),
		},
		"match_operator": &interop.AdvFunction{
			Name:    "match_operator",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("operator")},
			Value:   fuzzer.matchOperator,
		},
		"filter_codes": &tengo.UserFunction{
			Name:  "filter_codes",
			Value: fuzzer.funcASvRF(f.FilterCodes),
		},
		"filter_lines": &tengo.UserFunction{
			Name:  "filter_lines",
			Value: fuzzer.funcAIRF(f.FilterLines),
		},
		"filter_size": &tengo.UserFunction{
			Name:  "filter_size",
			Value: fuzzer.funcAIRF(f.FilterSize),
		},
		"filter_words": &tengo.UserFunction{
			Name:  "filter_words",
			Value: fuzzer.funcAIRF(f.FilterWords),
		},
		"filter_regex": &tengo.UserFunction{
			Name:  "filter_regex",
			Value: fuzzer.funcASRF(f.FilterRegex),
		},
		"filter_time": &tengo.UserFunction{
			Name:  "filter_time",
			Value: fuzzer.funcAIRF(f.FilterTime),
		},
		"filter_operator": &interop.AdvFunction{
			Name:    "filter_operator",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("operator")},
			Value:   fuzzer.filterOperator,
		},
		"authorization": &tengo.UserFunction{
			Name:  "authorization",
			Value: fuzzer.funcASRF(f.Authorization),
		},
		"bearer_token": &tengo.UserFunction{
			Name:  "bearer_token",
			Value: fuzzer.funcASRF(f.BearerToken),
		},
		"proxy": &tengo.UserFunction{
			Name:  "proxy",
			Value: fuzzer.funcASRF(f.Proxy),
		},
		"post_string": &tengo.UserFunction{
			Name:  "post_string",
			Value: fuzzer.funcASRF(f.PostString),
		},
		"post_json": &interop.AdvFunction{
			Name:    "post_json",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.ObjectArg("body")},
			Value:   fuzzer.postJSON,
		},
		"target": &tengo.UserFunction{
			Name:  "target",
			Value: fuzzer.funcASRF(f.Target),
		},
		"user_agent": &tengo.UserFunction{
			Name:  "user_agent",
			Value: fuzzer.funcASRF(f.UserAgent),
		},
		"http2": &tengo.UserFunction{
			Name:  "http2",
			Value: fuzzer.funcARF(f.HTTP2),
		},
		"ignore_body": &tengo.UserFunction{
			Name:  "ignore_body",
			Value: fuzzer.funcARF(f.IgnoreBody),
		},
		"follow_redirects": &tengo.UserFunction{
			Name:  "follow_redirects",
			Value: fuzzer.funcARF(f.FollowRedirects),
		},
		"dir_search_compat": &tengo.UserFunction{
			Name:  "dir_search_compat",
			Value: fuzzer.funcARF(f.DirSearchCompat),
		},
		"ignore_wordlist_comments": &tengo.UserFunction{
			Name:  "ignore_wordlist_comments",
			Value: fuzzer.funcARF(f.IgnoreWordlistComments),
		},
		"input_command": &tengo.UserFunction{
			Name:  "input_command",
			Value: fuzzer.funcASRF(f.InputCommand),
		},
		"input_num": &tengo.UserFunction{
			Name:  "input_num",
			Value: fuzzer.funcAIRF(f.InputNum),
		},
		"input_shell": &tengo.UserFunction{
			Name:  "input_shell",
			Value: fuzzer.funcASRF(f.InputShell),
		},
		"wordlist_mode": &interop.AdvFunction{
			Name:    "wordlist_mode",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("mode")},
			Value:   fuzzer.wordlistMode,
		},
		"raw_request_file": &tengo.UserFunction{
			Name:  "raw_request_file",
			Value: fuzzer.funcASRF(f.RawRequestFile),
		},
		"raw_request_protocol": &tengo.UserFunction{
			Name:  "raw_request_protocol",
			Value: fuzzer.funcASRF(f.RawRequestProtocol),
		},
		"wordlist": &tengo.UserFunction{
			Name:  "wordlist",
			Value: fuzzer.funcASRF(f.Wordlist),
		},
		"debug_log": &tengo.UserFunction{
			Name:  "debug_log",
			Value: fuzzer.funcASRF(f.DebugLog),
		},
		"output_file": &tengo.UserFunction{
			Name:  "output_file",
			Value: fuzzer.funcASRF(f.OutputFile),
		},
		"output_dir": &tengo.UserFunction{
			Name:  "output_dir",
			Value: fuzzer.funcASRF(f.OutputDir),
		},
		"output_format": &interop.AdvFunction{
			Name:    "output_format",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("format")},
			Value:   fuzzer.outputFormat,
		},
		"no_empty_output": &tengo.UserFunction{
			Name:  "no_empty_output",
			Value: fuzzer.funcARF(f.NoEmptyOutput),
		},
		"custom_arguments": &interop.AdvFunction{
			Name:    "custom_arguments",
			NumArgs: interop.MinArgs(1),
			Args:    []interop.AdvArg{interop.StrSliceArg("args", true)},
			Value:   fuzzer.customArguments,
		},
		"args": &tengo.UserFunction{
			Name:  "args",
			Value: stdlib.FuncARSs(f.Args),
		},
		"run": &tengo.UserFunction{
			Name:  "run",
			Value: stdlib.FuncARE(f.Run),
		},
		"run_with_output": &tengo.UserFunction{
			Name:  "run_with_output",
			Value: stdlib.FuncARSE(f.RunWithOutput),
		},
	}

	fuzzer.PropObject = types.PropObject{
		ObjectMap:  objectMap,
		Properties: make(map[string]types.Property),
	}

	return fuzzer
}

func newFfufFuzzer(s *Script) *Fuzzer {
	return makeFfufFuzzer(s, ffuf.NewFuzzer(s.ctx))
}
