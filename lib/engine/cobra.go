package engine

import (
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengomod/interop"
	"github.com/spf13/cobra"
)

// CobraModule represents the 'cobra' import module
func (s *Script) CobraModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"root_cmd": &interop.AdvFunction{
			Name:    "root_cmd",
			NumArgs: interop.ArgRange(1, 2),
			Args:    []interop.AdvArg{interop.StrArg("use"), interop.StrArg("short-description")},
			Value:   s.cobraRootCmd,
		},
		"cmd": &interop.AdvFunction{
			Name:    "cmd",
			NumArgs: interop.ArgRange(1, 2),
			Args:    []interop.AdvArg{interop.StrArg("use"), interop.StrArg("short-description")},
			Value:   s.cobraCmd,
		},
	}
}

func (s *Script) cobraRootCmd(args map[string]interface{}) (tengo.Object, error) {
	cmd, err := s.cobraCmd(args)
	if err != nil {
		return nil, err
	}

	c := cmd.(*CobraCmd)
	c.Value.SetArgs(s.args)
	c.Value.CompletionOptions.DisableDefaultCmd = true
	c.Value.SilenceErrors = true
	c.Value.SilenceUsage = true

	c.Value.Annotations["type"] = "root"

	c.Value.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		c.cobraRootCmdPersistentPreRun(cmd, args)
		return nil
	}

	return cmd, nil
}

func (s *Script) cobraCmd(args map[string]interface{}) (tengo.Object, error) {
	use := args["use"].(string)
	cmd := &cobra.Command{
		Use:         use,
		Annotations: map[string]string{"type": "sub"},
	}

	if value, ok := args["short-description"]; ok {
		cmd.Short = value.(string)
	}

	return makeCobraCmd(cmd, s), nil
}
