package capture

import (
	"crypto/md5"
	"fmt"
	builder "github.com/NoF0rte/cmd-builder"
	"github.com/analog-substance/arsenic/pkg/log"
	"github.com/analog-substance/util/fileutil"
	"github.com/charmbracelet/huh"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
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

func InteractiveRun(scopeDir string, cmdSlice []string, rerun bool) {
	cmdToRun := cmdSlice[0]
	cmdArgs := cmdSlice[1:]
	wrappedCmd := NewWrappedCommand(scopeDir, cmdToRun, cmdArgs)

	err := wrappedCmd.Run(rerun)
	if err != nil {
		logger.Error("failed to run command", "err", err)
	}
}

func injectOutputArgs(outputDir, cmd string, args []string, inputFile string) []string {
	newArgs := []string{}

	inputFileReplaced := false

	if cmd == "nmap" {
		for index, arg := range args {
			if strings.HasPrefix("-iL", arg) {
				args[index+1] = inputFile
				inputFileReplaced = true
				break
			}
		}

		if !inputFileReplaced {
			newArgs = append(newArgs, "-iL", inputFile)
		}

		//if !alreadyHasOutputFlag {
		logger.Debug("injecting output", "outputDir", outputDir)
		newArgs = append([]string{"-oA", outputDir}, newArgs...)
		//}
	} else if cmd == "subfinder" {
		for index, arg := range args {
			if strings.HasPrefix("-dL", arg) {
				args[index+1] = inputFile
				inputFileReplaced = true
				break
			}
		}

		if !inputFileReplaced {
			newArgs = append(newArgs, "-dL", inputFile)
		}

		newArgs = append(newArgs, "-json")
		newArgs = append([]string{"-oD", outputDir}, newArgs...)
		logger.Debug("injecting output", "outputDir", outputDir)
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

	input []string
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
			wc.Runs = append(wc.Runs, WrappedOutputFromCmdFile(file))
		}
	}
	return wc.Runs
}

func (wc *WrappedCommand) Run(rerun bool) error {
	if !rerun {
		shouldContinue, err := wc.MaybeAskToRerun()
		if err != nil {
			return err
		}
		if !shouldContinue {
			return fmt.Errorf("did not want to re-run command: %s %s", wc.Command, wc.Path)
		}
	}

	err := os.MkdirAll(wc.Path, 0755)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%d", time.Now().Unix())
	logger.Debug("capture file", "fileName", filename)

	var input []string
	if rerun {
		input = wc.getInput()
	} else {
		input = wc.GetInputNotUsedYet()
	}
	inputFile := wc.getPathFor(filename, "input")
	if len(input) > 0 {
		logger.Debug("input for command", "input", input)
		err = os.WriteFile(inputFile, []byte(strings.Join(input, "\n")+"\n"), 0755)
		if err != nil {
			return err
		}
	}

	cmdFilePath := wc.getPathFor(filename, "cmd")
	wc.Args = injectOutputArgs(filepath.Join(wc.Path, filename), wc.Command, wc.Args, inputFile)
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

func (wc *WrappedCommand) getInput() []string {
	if len(wc.input) == 0 {
		var inputList string
		var potentialTarget []string

		if wc.Command == "nmap" {
			singleTackFlags := `-(6|A|F|O|PE|PP|PM|PO|PS|PA|PU|PY|Pn|V|d|h|n|R|r|sC|sL|sN|sF|sX|sO|sS|sT|sA|sW|sM|sU|sV|sY|sZ|sn|v)`
			doubleTackFlags := `--(append\-output|badsum|iflist|noninteractive|no\-stylesheet|open|osscan\-guess|osscan\-limit|packet\-trace|privileged|reason|script\-args\-file|script\-trace|script\-updatedb|send\-eth|send\-ip|system\-dns|traceroute|unprivileged|version\-all|version\-light|version\-trace|webxml)`

			valuelessArgsRE := regexp.MustCompile(`^((` + singleTackFlags + `)|(` + doubleTackFlags + `))`)

			for index, arg := range wc.Args {
				if strings.HasPrefix("-iL", arg) {
					inputList = wc.Args[index+1]
					break
				}
				if index > 0 && !strings.HasPrefix(arg, "-") {
					// current arg is not a nmap flag/option
					// it could be a target specification
					if valuelessArgsRE.MatchString(wc.Args[index-1]) {
						// last arg was an option/switch that didn't need an argument
						potentialTarget = append(potentialTarget, arg)
					}
				}
			}
		} else if wc.Command == "radon" {
			for index, arg := range wc.Args {
				if strings.HasPrefix("-d", arg) || strings.HasPrefix("--domains-file", arg) {
					inputList = wc.Args[index+1]
					break
				}
			}
		} else if wc.Command == "subfinder" {
			for index, arg := range wc.Args {
				if strings.HasPrefix("-dL", arg) || strings.HasPrefix("--domains-file", arg) {
					inputList = wc.Args[index+1]
					break
				}
			}
		}

		if inputList != "" {
			logger.Debug("found input list", "cmd", wc.Command, "inputList", inputList)
			lines, err := fileutil.ReadLowerLines(inputList)
			if err != nil {
				logger.Error("failed to read input list", "err", err)
			} else {
				potentialTarget = append(potentialTarget, lines...)
			}
		}

		if len(potentialTarget) > 0 {
			logger.Debug("found potential targets", "cmd", wc.Command, "potentialTarget", potentialTarget)
		}

		targetMap := make(map[string]bool)
		for _, target := range potentialTarget {
			targetMap[target] = true
		}

		potentialTarget = []string{}
		for target := range targetMap {
			potentialTarget = append(potentialTarget, target)
		}

		sort.Strings(potentialTarget)

		wc.input = potentialTarget
	}

	return wc.input
}

func (wc *WrappedCommand) GetInputNotUsedYet() []string {
	inputAlreadyUsed := []string{}
	inputNotUsed := []string{}

	runs := wc.GetRuns()
	for _, input := range wc.getInput() {
		for _, run := range runs {
			if run.UsedInput(input) {
				inputAlreadyUsed = append(inputAlreadyUsed, input)
				break
			}
		}
		if !slices.Contains(inputAlreadyUsed, input) {
			inputNotUsed = append(inputNotUsed, input)
		}
	}

	logger.Debug("command already ran with input", "inputAlreadyUsed", inputAlreadyUsed, "inputNotUsed", inputNotUsed)
	return inputNotUsed
}

func (wc *WrappedCommand) MaybeAskToRerun() (bool, error) {
	if !fileutil.DirExists(wc.Path) {
		// file doesn't exist, lets roll!
		return true, nil
	}

	inputNotUsed := wc.GetInputNotUsedYet()
	if len(inputNotUsed) > 0 {
		// run with unused inputs
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
	Path            string
	StdErrFile      string
	StdOutFile      string
	AdditionalFiles []string

	input map[string]bool
}

func WrappedOutputFromCmdFile(cmdFile string) *WrappedOutput {

	dir := filepath.Dir(cmdFile)
	fileBase := filepath.Base(cmdFile)
	timeStampStr := strings.Split(fileBase, ".")[0]
	i, err := strconv.ParseInt(timeStampStr, 10, 64)
	if err != nil {
		panic(err)
	}
	timeStamp := time.Unix(i, 0)

	return &WrappedOutput{
		Path:    dir,
		Created: timeStamp,
	}
}

func (wo *WrappedOutput) GetInput() (map[string]bool, error) {
	if len(wo.input) == 0 {
		inputFile := filepath.Join(wo.Path, fmt.Sprintf("%d.input", wo.Created.Unix()))
		lines, err := fileutil.ReadLowerLineMap(inputFile)
		if err != nil {
			logger.Debug("error loading input", "inputFile", inputFile, "err", err)
			return nil, err
		}
		wo.input = lines
	}
	return wo.input, nil
}

func (wo *WrappedOutput) UsedInput(input string) bool {

	inputMap, err := wo.GetInput()
	logger.Debug("run used input?", "input", input, "currentInput", inputMap)
	if err != nil {
		return false
	}

	if _, ok := inputMap[input]; ok {
		return true
	}
	return false
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
