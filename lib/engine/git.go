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

var gitModule *GitModule = &GitModule{
	isGit: util.DirExists(".git"),
}

type GitModule struct {
	isGit     bool
	moduleMap map[string]tengo.Object
}

func (m *GitModule) ModuleMap() map[string]tengo.Object {
	if m.moduleMap == nil {
		m.moduleMap = map[string]tengo.Object{
			"pull":   &tengo.UserFunction{Name: "pull", Value: m.tengoPull},
			"commit": &tengo.UserFunction{Name: "commit", Value: m.tengoCommit},
			"lock":   &tengo.UserFunction{Name: "lock", Value: m.tengoLock},
		}
	}
	return m.moduleMap
}

func (m *GitModule) tengoPull(args ...tengo.Object) (tengo.Object, error) {
	if !m.isGit {
		return nil, nil
	}

	err := m.pull(true)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

func (m *GitModule) pull(rebase bool) error {
	if !m.isGit {
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

func (m *GitModule) tengoCommit(args ...tengo.Object) (tengo.Object, error) {
	if !m.isGit {
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

	err := m.commit(path, message, mode)
	if err != nil {
		return toError(err), nil
	}
	return nil, nil
}

func (m *GitModule) hardReset() {
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

func (m *GitModule) commit(path string, msg string, mode string) error {
	if !m.isGit {
		return nil
	}

	err := m.add(path)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		fmt.Println("nothing happened")
		return nil
	}

	err = m.push()
	if err == nil {
		return nil
	}

	util.LogWarn("First push failed", err)

	err = m.pull(true)
	if err != nil {
		util.LogWarn("pull rebase failed", err)
		if mode == "reset" {
			m.hardReset()
		}
		os.Exit(2)
	}

	return m.push()
}

func (m *GitModule) push() error {
	return exec.Command("git", "push").Run()
}

func (m *GitModule) add(path string) error {
	return exec.Command("git", "add", path).Run()
}

func (m *GitModule) lock(lockFile string, msg string) error {
	if !m.isGit {
		return nil
	}

	if util.FileExists(lockFile) {
		util.LogWarn("can't lock a file that exists")
		scriptModule.stop()
		return nil
	}

	err := m.pull(true)
	if err != nil {
		m.hardReset()
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

	return m.commit(lockFile, msg, "reset")
}

func (m *GitModule) tengoLock(args ...tengo.Object) (tengo.Object, error) {
	if !m.isGit {
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

	err := m.lock(lockFile, msg)
	if err != nil {
		return toError(err), nil
	}

	return nil, nil
}
