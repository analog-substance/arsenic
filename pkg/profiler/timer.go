package profiler

import (
	"github.com/analog-substance/arsenic/pkg/log"
	"log/slog"
	"time"
)

var logger *slog.Logger

func init() {
	logger = log.WithGroup("time")
}

func Timer(name string) func() {
	start := time.Now()
	return func() {
		logger.Info("completed", "name", name, "duration", time.Since(start))
	}
}
