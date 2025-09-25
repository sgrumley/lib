package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/lmittmann/tint"
	"github.com/phsym/console-slog"
)

type Logger struct {
	*slog.Logger
}

type Level slog.Level

const (
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelError Level = Level(slog.LevelError)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelDebug Level = Level(slog.LevelDebug)
)

type Handler string

const (
	HandlerJSON    Handler = "json"
	HandlerText    Handler = "text"
	HandlerConsole Handler = "console"
	HandlerDevsLog Handler = "devslog"
	HandlerTint    Handler = "tint"
)

type (
	Option        func(*LoggerOptions)
	LoggerOptions struct {
		level  Level
		format Handler
		output *os.File
		source bool
	}
)

func WithLevel(level Level) Option {
	return func(opts *LoggerOptions) {
		opts.level = level
	}
}

func WithFormat(format Handler) Option {
	return func(opts *LoggerOptions) {
		opts.format = format
	}
}

func WithSource(s bool) Option {
	return func(opts *LoggerOptions) {
		opts.source = s
	}
}

func WithOutput(out *os.File) Option {
	return func(opts *LoggerOptions) {
		opts.output = out
	}
}

func NewLogger(options ...Option) *Logger {
	// default
	opts := LoggerOptions{
		level:  LevelInfo,
		format: HandlerText,
		output: os.Stdout,
		source: false,
	}

	for _, opt := range options {
		opt(&opts)
	}

	handlerPreset := getHandler(opts)
	logger := slog.New(handlerPreset)
	// this allows access via importing slog, however it is better to pass
	// 	the logger where you can to avoid modifying the global instance
	slog.SetDefault(logger)
	return &Logger{
		slog.New(handlerPreset),
	}
}

func getHandler(opts LoggerOptions) slog.Handler {
	baseOpts := &slog.HandlerOptions{
		AddSource: opts.source,
		Level:     slog.Level(opts.level),
	}

	switch opts.format {
	case HandlerJSON:
		return slog.NewJSONHandler(opts.output, baseOpts)

	case HandlerText:
		return slog.NewTextHandler(opts.output, baseOpts)

	case HandlerConsole:
		return console.NewHandler(opts.output, &console.HandlerOptions{
			AddSource: opts.source,
			Level:     slog.Level(opts.level),
			Theme:     console.NewBrightTheme(),
		})

	case HandlerDevsLog:
		return devslog.NewHandler(opts.output, &devslog.Options{
			HandlerOptions:    baseOpts,
			MaxSlicePrintSize: 4,
			SortKeys:          false,
			TimeFormat:        "[04:05]",
			NewLineAfterLog:   true,
			DebugColor:        devslog.Magenta,
			InfoColor:         devslog.Green,
			ErrorColor:        devslog.Red,
			WarnColor:         devslog.Yellow,
		})

	case HandlerTint:
		w := os.Stderr
		return tint.NewHandler(w, &tint.Options{
			AddSource:  false,
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			NoColor:    false,
		})
	}

	return slog.NewTextHandler(opts.output, baseOpts)
}

func (l *Logger) Error(msg string, err error) {
	l.Logger.Error(msg, slog.Any("error", err))
}

func (l *Logger) Fatal(msg string, err error) {
	l.Error(msg, err)
	os.Exit(1)
}

func (l *Logger) With(args ...any) *Logger {
	lw := l.Logger.With(args)
	return &Logger{
		lw,
	}
}

/*
usage:
binding records to logger:
logger = logger.With("user_id", 123)

grouping records:
logger.Info("a test message",
    slog.Group("user",
        slog.Int("user_id", 123),
        slog.String("user_name", "John Doe"),
    ),
    slog.Group("account",
        "money", 1000000,
        slog.String("type", "premium"),
    ),
)

logging structs:
logger.Info("checkout", "user", user)

custom struct logging:
type LogValuer interface {
	LogValue() Value
}
// this would alter how User is logged
func (u User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
        slog.String("password", "******"),
	)
}

logger.Info("checkout", "user", User{ID: 123, Name: "John Doe", Password: "123456"})
*/
