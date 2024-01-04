package engine

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/analog-substance/arsenic/lib/log"
	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengomod/interop"
)

func (s *Script) GitModule() map[string]tengo.Object {
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
		return interop.GoErrToTErr(err), nil
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

	err := runWithError(exec.Command("git", args...))
	if err != nil {
		log.Warn("pull failed")
		return err
	}

	return nil
}

func (s *Script) tengoCommit(args ...tengo.Object) (tengo.Object, error) {
	if !s.isGit {
		return nil, nil
	}

	if len(args) > 2 && len(args) <= 3 {
		return nil, tengo.ErrWrongNumArguments
	}

	path, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	message, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "message",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	mode := ""
	if len(args) == 3 {
		mode, ok = tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "mode",
				Expected: "string",
				Found:    args[2].TypeName(),
			}
		}
	}

	err := s.commit(path, message, mode)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}
	return nil, nil
}

func (s *Script) hardReset() {
	log.Warn("reset to origin")

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

	// Might want this to be on a loop with a maximum number of attempts.
	// This will fail if something is checked after we do a rebase pull.
	// Maybe we "parse" the error to know whether it is a true fail or
	// we need to do a rebase pull
	err = s.push()
	if err == nil {
		return nil
	}

	log.Warn("First push failed: ", err)

	err = s.pull(true)
	if err != nil {
		log.Warn("pull rebase failed: ", err)
		if mode == "reset" {
			s.hardReset()
		}
		os.Exit(2)
	}

	return s.push()
}

func (s *Script) push() error {
	if !s.isGit {
		return nil
	}

	return runWithError(exec.Command("git", "push"))
}

func (s *Script) add(path string) error {
	if !s.isGit {
		return nil
	}

	return runWithError(exec.Command("git", "add", path))
}

func (s *Script) lock(lockFile string, msg string) error {
	if fileutil.FileExists(lockFile) {
		log.Warn("can't lock a file that exists")
		s.fatal("")
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
	err = fileutil.WriteString(lockFile, content)
	if err != nil {
		return err
	}

	return s.commit(lockFile, msg, "reset")
}

func (s *Script) tengoLock(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	lockFile, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "lockFile",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	msg, ok := tengo.ToString(args[1])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "msg",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	err := s.lock(lockFile, msg)
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return nil, nil
}
