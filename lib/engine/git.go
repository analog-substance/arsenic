package engine

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/analog-substance/arsenic/lib/util"
	"github.com/d5/tengo/v2"
)

func (s *Script) GitModuleMap() map[string]tengo.Object {
	return map[string]tengo.Object{
		"pull":   &tengo.UserFunction{Name: "pull", Value: s.tengoPull},
		"commit": &tengo.UserFunction{Name: "commit", Value: s.tengoCommit},
		"lock":   &tengo.UserFunction{Name: "lock", Value: s.tengoLock},
	}
}

func (s *Script) tengoPull(args ...tengo.Object) (tengo.Object, error) {
	if !s.isGit {
		return nil, nil
	}

	err := s.pull(true)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

func (s *Script) pull(rebase bool) error {
	if !s.isGit {
		return nil
	}
	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	}

	cmd := exec.Command("git", args...)
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return err
	}

	errorText := ""
	go func() {
		buf := new(bytes.Buffer)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			buf.WriteString(scanner.Text() + "\n")
		}

		errorText = buf.String()
	}()

	err := cmd.Wait()
	if err != nil {
		util.LogWarn("pull failed")
		return errors.New(errorText)
	}

	return nil
}

func (s *Script) tengoCommit(args ...tengo.Object) (tengo.Object, error) {
	if !s.isGit {
		return nil, nil
	}

	if len(args) > 2 && len(args) <= 3 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	message, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "message",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
	}

	mode := ""
	if len(args) == 3 {
		mode, ok = tengo.ToString(args[2])
		if !ok {
			return toError(tengo.ErrInvalidArgumentType{
				Name:     "mode",
				Expected: "string",
				Found:    args[2].TypeName(),
			}), nil
		}
	}

	err := s.commit(path, message, mode)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

func (s *Script) hardReset() {
	util.LogWarn("reset to origin")

	rebaseCmd := exec.Command("git", "rebase", "--abort")
	rebaseCmd.Stdout = os.Stdout
	rebaseCmd.Stderr = os.Stderr

	rebaseCmd.Run()

	resetCmd := exec.Command("git", "reset", "--hard", "origin/master")
	resetCmd.Stdout = os.Stdout
	resetCmd.Stderr = os.Stderr

	resetCmd.Run()
}

func (s *Script) commit(path string, msg string, mode string) error {
	if !s.isGit {
		return nil
	}

	err := s.add(path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "commit", "-m", msg)
	err = cmd.Run()
	if err != nil {
		fmt.Println("nothing happened")
		return nil
	}

	err = s.push()
	if err == nil {
		return nil
	}

	util.LogWarn("First push failed", err)

	err = s.pull(true)
	if err != nil {
		util.LogWarn("pull rebase failed", err)
		if mode == "reset" {
			s.hardReset()
		}
		os.Exit(2)
	}

	return s.push()
}

func (s *Script) push() error {
	return exec.Command("git", "push").Run()
}

func (s *Script) add(path string) error {
	return exec.Command("git", "add", path).Run()
}

func (s *Script) lock(lockFile string, msg string) error {
	if !s.isGit {
		return nil
	}

	if util.FileExists(lockFile) {
		util.LogWarn("can't lock a file that exists")
		s.stop()
		return nil
	}

	err := s.pull(true)
	if err != nil {
		s.hardReset()
		os.Exit(1)
	}

	r := rand.Reader
	b := make([]byte, 16)
	_, _ = r.Read(b)

	content := fmt.Sprintf("lock::%s", base64.RawURLEncoding.EncodeToString(b))
	err = os.WriteFile(lockFile, []byte(content), util.DefaultFilePerms)
	if err != nil {
		return err
	}

	return s.commit(lockFile, msg, "reset")
}

func (s *Script) tengoLock(args ...tengo.Object) (tengo.Object, error) {
	if !s.isGit {
		return nil, nil
	}

	if len(args) != 2 {
		return toError(tengo.ErrWrongNumArguments), nil
	}

	lockFile, ok := tengo.ToString(args[0])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "lockFile",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}

	msg, ok := tengo.ToString(args[1])
	if !ok {
		return toError(tengo.ErrInvalidArgumentType{
			Name:     "msg",
			Expected: "string",
			Found:    args[1].TypeName(),
		}), nil
	}

	err := s.lock(lockFile, msg)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}