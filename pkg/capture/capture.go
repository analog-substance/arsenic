package capture

import (
	"crypto/md5"
	"fmt"
	builder "github.com/NoF0rte/cmd-builder"
	"github.com/analog-substance/arsenic/pkg/log"
	"github.com/analog-substance/fileutil"
	"github.com/charmbracelet/huh"
	"io"
	"log/slog"
	"sort"

	"os"
	"path"
	"strings"
	"time"
)

var logger *slog.Logger

func init() {
	logger = log.WithGroup("capture")
}

func InteractiveRun(scopeDir string, cmdSlice []string) {

	cmdToRun := cmdSlice[0]
	cmdArgs := cmdSlice[1:]

	cmdBase := path.Base(cmdToRun)
	md5Sum := md5.Sum([]byte(strings.Join(cmdArgs, " ")))
	argsHash := fmt.Sprintf("%x", md5Sum)

	outputDir := path.Join("data", scopeDir, "output", cmdBase, argsHash)

	if fileutil.DirExists(outputDir) {
		var confirm bool
		logger.Debug("already exists")

		err := huh.NewConfirm().
			Title("Looks like you already ran this").
			Description(getLatestFile(outputDir)).
			Affirmative("Yes!").
			Negative("No.").
			Value(&confirm).Run()

		if !confirm || err != nil {
			return
		}
	}

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	//re := regexp.MustCompile(`[^a-zA-Z0-9_\-.]+`)
	filename := fmt.Sprintf("%d", time.Now().Unix())

	logger.Debug("capture file", "fileName", filename)
	stdOutFilename := fmt.Sprintf("%s.stdout", filename)
	stdErrFilename := fmt.Sprintf("%s.stderr", filename)
	cmdFilename := fmt.Sprintf("%s.cmd", filename)

	stdOutFilePath := path.Join(outputDir, stdOutFilename)
	stdErrFilePath := path.Join(outputDir, stdErrFilename)
	cmdFilePath := path.Join(outputDir, cmdFilename)

	cmdArgs = injectOutputArgs(path.Join(outputDir, filename), cmdToRun, cmdArgs)
	cmdFileContents := fmt.Sprintf("%s %s", cmdToRun, strings.Join(cmdArgs, " "))
	err = os.WriteFile(cmdFilePath, []byte(cmdFileContents), 0755)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	stdOutFile, err := os.OpenFile(stdOutFilePath, os.O_WRONLY^os.O_CREATE, 0644)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer stdOutFile.Close()

	stdErrFile, err := os.OpenFile(stdErrFilePath, os.O_WRONLY^os.O_CREATE, 0644)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer stdErrFile.Close()

	outMultiWriter := io.MultiWriter(os.Stdout, stdOutFile)
	errMultiWriter := io.MultiWriter(os.Stderr, stdErrFile)
	wrapped := builder.Cmd(cmdToRun, cmdArgs...).
		Stdin(os.Stdin).
		Stdout(outMultiWriter).
		Stderr(errMultiWriter).
		Build()

	logger.Debug("wrapped command", "cmdToRun", cmdToRun, "cmdArgs", cmdArgs)

	err = wrapped.Run()
	if err != nil {
		logger.Error(err.Error())
	}

}

func injectOutputArgs(outputFile, cmd string, args []string) []string {
	newArgs := []string{}

	alreadyHasOutputFlag := false

	if cmd == "nmap" {
		for _, arg := range args {
			if strings.HasPrefix("-o", arg) {
				alreadyHasOutputFlag = true
				break
			}
		}

		if !alreadyHasOutputFlag {
			logger.Debug("injecting output", "outputFile", outputFile)
			newArgs = append([]string{"-oA", outputFile}, newArgs...)
		}
	}
	return append(newArgs, args...)
}

func getLatestFile(dir string) string {
	files := []string{}
	dirListing, err := os.ReadDir(dir)
	if err != nil {
		logger.Error(err.Error())
	}
	for _, file := range dirListing {
		if strings.HasSuffix(file.Name(), ".stdout") {
			files = append(files, file.Name())
		}
	}
	sort.Strings(files)
	bytes, err := os.ReadFile(path.Join(dir, files[0]))
	if err != nil {
		logger.Error(err.Error())
	}
	return string(bytes)

}
