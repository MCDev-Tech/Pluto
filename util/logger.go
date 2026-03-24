package util

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
)

type SlogWriter struct {
	Level slog.Level
}

func (s *SlogWriter) Write(p []byte) (n int, err error) {
	Logger.Log(context.Background(), s.Level, strings.Trim(string(p), "\n\r "))
	return len(p), nil
}

var Logger = slog.New(tint.NewHandler(os.Stderr, &tint.Options{
	Level:      slog.LevelDebug,
	TimeFormat: time.Kitchen,
	NoColor:    false,
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.MessageKey {
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue("\x1B[1m" + a.Value.String() + "\033[0m"),
			}
		}
		return a
	},
}))

func InitLogger() {
	slog.SetDefault(Logger)
}
