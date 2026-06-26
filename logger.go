package wzlib

import (
	"log/slog"
	"os"
)

func SetLogger(level slog.Level, addSource bool) {
	logLevelVar := new(slog.LevelVar)
	logLevelVar.Set(level)
	loggerHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     logLevelVar,
		AddSource: addSource,
	})
	logger := slog.New(loggerHandler)
	slog.SetDefault(logger)
}
