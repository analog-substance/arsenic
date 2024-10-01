package capture

import (
	"fmt"
	builder "github.com/NoF0rte/cmd-builder"
	"github.com/analog-substance/arsenic/pkg/log"
	"io"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var logger *slog.Logger

func init() {
	logger = log.WithGroup("capture")
}

func Run(scopeDir string, cmdSlice []string) {

	cmdToRun := cmdSlice[0]
	cmdArgs := cmdSlice[1:]

	cmdBase := path.Base(cmdToRun)
	outputDir := path.Join("data", scopeDir, "output", cmdBase)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9_\-.]+`)

	fileName := re.ReplaceAllString(strings.Join(cmdArgs, "_"), "__")

	logger.Debug("capture file", "fileName", fileName)
	outputFileName := fmt.Sprintf("%d__%s", time.Now().Unix(), fileName)
	cmdFileName := fmt.Sprintf("%d__%s.cmd", time.Now().Unix(), fileName)

	outputFilePath := path.Join(outputDir, outputFileName)
	cmdFilePath := path.Join(outputDir, cmdFileName)

	cmdArgs = injectOutputArgs(outputFileName, cmdToRun, cmdArgs)

	cmdFilecontents := fmt.Sprintf("%s%s", cmdToRun, strings.Join(cmdArgs, " "))
	err = os.WriteFile(cmdFilePath, []byte(cmdFilecontents), 0755)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	file, err := os.OpenFile(outputFilePath, os.O_WRONLY^os.O_CREATE, 0644)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer file.Close()
	multiWriter := io.MultiWriter(os.Stdout, file)
	wrapped := builder.Cmd(cmdToRun, cmdArgs[1:]...).
		Stdin(os.Stdin).
		Stdout(multiWriter).
		Build()

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
			newArgs = append(newArgs, "-oA", outputFile)
		}
	}
	return append(newArgs, args...)
}
