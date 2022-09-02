package script

import (
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
			"pull":   &tengo.UserFunction{Name: "pull", Value: m.pull},
			"commit": &tengo.UserFunction{Name: "commit", Value: m.commit},
			"lock":   &tengo.UserFunction{Name: "lock", Value: m.lock},
		}
	}
	return m.moduleMap
}

func (m *GitModule) pull(args ...tengo.Object) (tengo.Object, error) {
	if m.isGit {
		cmd := exec.Command("git", "pull", "--rebase")

		err := cmd.Run()
		if err != nil {
			util.LogWarn("pull failed")
		}
	}
	return nil, nil
}
func (m *GitModule) commit(args ...tengo.Object) (tengo.Object, error) {
	return nil, nil
}
func (m *GitModule) lock(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	m.pull()

	lockFile, _ := tengo.ToString(args[0])
	msg, _ := tengo.ToString(args[1])
	if util.FileExists(lockFile) {
		util.LogWarn("can't lock a file that exists")
		os.Exit(1)
	}

	// TODO: Write random base64 string to lock file

	gitCommit(lockFile, msg, "reset")

	return nil, nil
}

func gitCommit(path string, msg string, mode string) error {
	if !gitModule.isGit {
		return nil
	}

	return nil
}
