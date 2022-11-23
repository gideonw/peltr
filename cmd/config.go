package cmd

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func configLog(app string) zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).
		With().
		Timestamp().
		Caller().
		Str("app", app).
		Logger()

	return logger
}
