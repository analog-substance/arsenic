package engine

import (
	"github.com/analog-substance/tengo/v2"
	"github.com/spf13/cobra"
)

func (s *Script) CobraModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"root_cmd": &tengo.UserFunction{Name: "root_cmd", Value: s.cobraRootCmd},
		"cmd":      &tengo.UserFunction{Name: "cmd", Value: s.cobraCmd},
	}
}

func (s *Script) cobraRootCmd(args ...tengo.Object) (tengo.Object, error) {
	cmd, err := s.cobraCmd(args...)
	if err != nil {
		return nil, err
	}

	(cmd.(*CobraCmd)).Value.SetArgs(s.args)
	return cmd, nil
}

func (s *Script) cobraCmd(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 || len(args) >= 3 {
		return nil, tengo.ErrWrongNumArguments
	}

	use, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "use",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	cmd := &cobra.Command{
		Use: use,
	}

	if len(args) == 2 {
		shortDesc, ok := tengo.ToString(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "short description",
				Expected: "string",
				Found:    args[1].TypeName(),
			}
		}
		cmd.Short = shortDesc
	}

	return makeCobraCmd(cmd, s), nil
}
