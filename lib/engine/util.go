package engine

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/analog-substance/tengo/v2"
)

// toWarning creates a Warning object from a string
func toWarning(value string) tengo.Object {
	return &Warning{
		Value: &tengo.String{
			Value: value,
		},
	}
}

func runWithError(cmd *exec.Cmd) error {
	buf := new(bytes.Buffer)
	cmd.Stderr = buf

	if err := cmd.Start(); err != nil {
		return err
	}

	err := cmd.Wait()
	if err != nil {
		return errors.New(buf.String())
	}
	return nil
}
