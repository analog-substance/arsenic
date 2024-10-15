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
	"path/filepath"
	"sort"
	"strconv"

	"os"
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
	wrappedCmd := NewWrappedCommand(scopeDir, cmdToRun, cmdArgs)

	err := wrappedCmd.Run()
	if err != nil {
		logger.Error("failed to run command", "err", err)
	}
}

func injectOutputArgs(outputFile, cmd string, args []string) []string {
	newArgs := []string{}

	//alreadyHasOutputFlag := false

	if cmd == "nmap" {
		//for _, arg := range args {
		//	if strings.HasPrefix("-o", arg) {
		//		alreadyHasOutputFlag = true
		//		break
		//	}
		//}

		//if !alreadyHasOutputFlag {
		logger.Debug("injecting output", "outputFile", outputFile)
		newArgs = append([]string{"-oA", outputFile}, newArgs...)
		//}
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
	bytes, err := os.ReadFile(filepath.Join(dir, files[0]))
	if err != nil {
		logger.Error(err.Error())
	}
	return string(bytes)
}

type WrappedCommand struct {
	Path    string
	Command string
	Args    []string
	Runs    []*WrappedOutput
}

func GetWrappedCommands(scopeDir string) []*WrappedCommand {

	wrappedCommands := []*WrappedCommand{}

	outputDir := filepath.Join("data", scopeDir, "output")
	cmdDirs, err := os.ReadDir(outputDir)
	if err != nil {
		logger.Error(err.Error())
		return wrappedCommands
	}

	for _, cmdDir := range cmdDirs {
		if cmdDir.IsDir() {
			commandName := cmdDir.Name()
			runs, err := os.ReadDir(filepath.Join(outputDir, commandName))
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			for _, run := range runs {
				cmdFiles, err := filepath.Glob(filepath.Join(outputDir, cmdDir.Name(), run.Name(), "*.cmd"))
				if err != nil {
					logger.Error(err.Error())
					continue
				}
				for _, cmdFile := range cmdFiles {
					fileBytes, err := os.ReadFile(cmdFile)
					if err != nil {
						logger.Error(err.Error())
						continue
					}

					cmdLine := strings.Split(string(fileBytes), " ")

					wrappedCommands = append(wrappedCommands, NewWrappedCommand(scopeDir, cmdDir.Name(), cmdLine[1:]))
					break
				}
			}
		}
	}

	return wrappedCommands
}

func (wc *WrappedCommand) GetRuns() []*WrappedOutput {
	if len(wc.Runs) == 0 {
		files, err := filepath.Glob(fmt.Sprintf("%s/*.cmd", wc.Path))
		if err != nil {
			logger.Error(err.Error())
		}

		for _, file := range files {
			wc.Runs = append(wc.Runs, WrappedOutputFromCmd(filepath.Join(wc.Path, file)))
		}
	}
	return wc.Runs
}

func (wc *WrappedCommand) Run() error {

	shouldContine, err := wc.MaybeAskToRerun()
	if err != nil {
		return err
	}
	if !shouldContine {
		return fmt.Errorf("did not want to re-run command: %s %s", wc.Command, wc.Path)
	}

	err = os.MkdirAll(wc.Path, 0755)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%d", time.Now().Unix())
	logger.Debug("capture file", "fileName", filename)

	cmdFilePath := wc.getPathFor(filename, "cmd")
	wc.Args = injectOutputArgs(filepath.Join(wc.Path, filename), wc.Command, wc.Args)
	cmdFileContents := fmt.Sprintf("%s %s", wc.Command, strings.Join(wc.Args, " "))
	err = os.WriteFile(cmdFilePath, []byte(cmdFileContents), 0755)
	if err != nil {
		return err
	}

	stdOutFile, err := wc.getOutputFileWriter(filename, "stdout")
	if err != nil {
		return err
	}
	defer stdOutFile.Close()

	stdErrFile, err := wc.getOutputFileWriter(filename, "stderr")
	if err != nil {
		return err
	}
	defer stdErrFile.Close()

	outMultiWriter := io.MultiWriter(os.Stdout, stdOutFile)
	errMultiWriter := io.MultiWriter(os.Stderr, stdErrFile)
	wrapped := builder.Cmd(wc.Command, wc.Args...).
		Stdin(os.Stdin).
		Stdout(outMultiWriter).
		Stderr(errMultiWriter).
		Build()

	logger.Debug("wrapped command", "cmdToRun", wc.Args, "cmdArgs", wc.Args)

	return wrapped.Run()
}

func (wc *WrappedCommand) MaybeAskToRerun() (bool, error) {
	if !fileutil.DirExists(wc.Path) {
		// file doesn't exist, lets roll!
		return true, nil
	}

	// file exists, lets ask what we should do.
	var confirm bool
	logger.Debug("already exists")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Previous command Output").
				Description(getLatestFile(wc.Path)),

			huh.NewConfirm().
				Title("Looks like you already ran this").
				//Description(getLatestFile(outputDir)).
				Affirmative("Yes!").
				Negative("No.").
				Value(&confirm),
		),
	)

	err := form.Run()
	if err != nil {
		return false, err
	}

	return confirm, nil

}

type WrappedOutput struct {
	Created         time.Time
	StdErrFile      string
	StdOutFile      string
	AdditionalFiles []string
}

func WrappedOutputFromCmd(cmdFile string) *WrappedOutput {

	fileBase := filepath.Base(cmdFile)
	timeStampStr := strings.Split(fileBase, ".")[0]
	i, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		panic(err)
	}
	timeStamp := time.Unix(i, 0)

	return &WrappedOutput{
		Created: timeStamp,
	}
}

func NewWrappedCommandFromDir(cmdDir string) *WrappedCommand {
	cmdName, cmdArgs := getCmd(cmdDir)

	return &WrappedCommand{
		Path:    cmdDir,
		Command: cmdName,
		Args:    cmdArgs,
	}
}

func NewWrappedCommand(scopeDir, cmdName string, cmdArgs []string) *WrappedCommand {

	cmdBase := filepath.Base(cmdName)
	md5Sum := md5.Sum([]byte(strings.Join(cmdArgs, " ")))
	argsHash := fmt.Sprintf("%x", md5Sum)
	outputDir := filepath.Join("data", scopeDir, "output", cmdBase, argsHash)

	return &WrappedCommand{
		Path:    outputDir,
		Command: cmdName,
		Args:    cmdArgs,
	}
}

func getCmd(dir string) (string, []string) {
	dirListing, err := os.ReadDir(dir)
	if err != nil {
		logger.Error(err.Error())
	}
	for _, file := range dirListing {
		if strings.HasSuffix(file.Name(), ".cmd") {
			bytes, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				logger.Error(err.Error())
			}

			commandLine := strings.Split(string(bytes), " ")
			return commandLine[0], commandLine[1:]
		}
	}
	return "", []string{}
}

func (wc *WrappedCommand) getPathFor(filename, suffix string) string {
	newFilename := fmt.Sprintf("%s.%s", filename, suffix)
	return filepath.Join(wc.Path, newFilename)
}

func (wc *WrappedCommand) getOutputFileWriter(filename, suffix string) (*os.File, error) {
	outputFilePath := wc.getPathFor(filename, suffix)
	return os.OpenFile(outputFilePath, os.O_WRONLY^os.O_CREATE, 0644)
}
