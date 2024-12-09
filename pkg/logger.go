package pkg

import (
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	logFile     *os.File
	LogFileName string
}

func (l *Logger) InitLogger() {
	var w io.Writer

	if l.LogFileName != "" {
		logFile, err := os.OpenFile(l.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		l.logFile = logFile

		w = io.MultiWriter(os.Stdout, l.logFile)
	} else {
		w = io.MultiWriter(os.Stdout)

	}

	newTextHandler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(newTextHandler)
	slog.SetDefault(logger)
}
