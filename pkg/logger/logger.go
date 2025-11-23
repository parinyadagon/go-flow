package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(environment string) {
	var output io.Writer = os.Stdout

	// Pretty console output for development
	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if environment == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = Logger
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
