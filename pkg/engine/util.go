package engine

import (
	"bytes"
	"errors"
	"os/exec"
)

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
