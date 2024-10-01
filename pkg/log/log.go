package log

import (
	"fmt"
	"log/slog"
	"os"
)

var logLevel = new(slog.LevelVar)
var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
}

func Logger() *slog.Logger {
	return logger
}

func WithGroup(groupName string) *slog.Logger {
	return logger.With(slog.Group("carbon", slog.String("pkg", groupName)))
}

func LogLevel(level slog.Level) {
	logLevel.Set(level)
}

func Msg(args ...interface{}) {
	log("[+]", args...)
}
func Warn(args ...interface{}) {
	log("[!]", args...)
}
func Info(args ...interface{}) {
	log("[-]", args...)
}

func log(prefix string, args ...interface{}) {
	fmt.Printf("%s ", prefix)
	fmt.Print(args...)
	fmt.Println()
}
