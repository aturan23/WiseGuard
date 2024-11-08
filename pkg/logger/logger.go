package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Config logger configuration
type Config struct {
	Level  string // log level: debug, info, warn, error, fatal
	Pretty bool   // pretty print
}

type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
	Fatal(msg string, err error, fields map[string]interface{})
	WithComponent(component string) Logger
	WithFields(fields map[string]interface{}) Logger
}

type ZeroLogger struct {
	log zerolog.Logger
}

func NewLogger(config Config) Logger {
	// Set output
	var output io.Writer = os.Stdout
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()

	return &ZeroLogger{
		log: logger,
	}
}

func (l *ZeroLogger) Debug(msg string, fields map[string]interface{}) {
	event := l.log.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *ZeroLogger) Info(msg string, fields map[string]interface{}) {
	event := l.log.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *ZeroLogger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.log.Error()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *ZeroLogger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.log.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

func (l *ZeroLogger) WithComponent(component string) Logger {
	return &ZeroLogger{
		log: l.log.With().Str("component", component).Logger(),
	}
}

func (l *ZeroLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.log.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &ZeroLogger{
		log: ctx.Logger(),
	}
}
