package script

import (
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

	cmd := exec.Command("git", "pull", "--rebase")

	err := cmd.Run()
	if err != nil {
		util.LogWarn("pull failed")
	}
	return nil, err
}

func (m *GitModule) tengoCommit(args ...tengo.Object) (tengo.Object, error) {
	if !m.isGit {
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

	return nil, m.commit(path, message, mode)
}

func (m *GitModule) commit(path string, msg string, mode string) error {
	if !m.isGit {
		return nil
	}

	err := m.add(path)
	if err != nil {
		return err
	}

	err = exec.Command("git", "commit", "-m", msg).Run()
	if err != nil {
		fmt.Println("nothing happened")
		return nil
	}

	err = m.push()
	if err == nil {
		return nil
	}

	util.LogWarn("First push failed", err)

	err = exec.Command("git", "pull", "--rebase").Run()
	if err != nil {
		util.LogWarn("pull rebase failed", err)
		if mode == "reset" {
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

func (m *GitModule) tengoLock(args ...tengo.Object) (tengo.Object, error) {
	if !m.isGit {
		return nil, nil
	}

	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	m.tengoPull()

	lockFile, _ := tengo.ToString(args[0])
	msg, _ := tengo.ToString(args[1])
	if util.FileExists(lockFile) {
		util.LogWarn("can't lock a file that exists")
		stopScript()
	}

	// TODO: Write random base64 string to lock file

	m.commit(lockFile, msg, "reset")

	return nil, nil
}
