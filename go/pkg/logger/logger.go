package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Options struct {
	Level     string
	Format    Format
	Output    io.Writer
	AddSource bool
	Attrs     []slog.Attr
}

func New(opts Options) (*slog.Logger, error) {
	level, err := ParseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: opts.AddSource,
	}

	var handler slog.Handler
	switch Format(strings.ToLower(string(opts.Format))) {
	case FormatJSON:
		handler = slog.NewJSONHandler(out, handlerOpts)
	case FormatText, "":
		handler = slog.NewTextHandler(out, handlerOpts)
	default:
		return nil, fmt.Errorf("unknown log format %q", opts.Format)
	}

	if len(opts.Attrs) > 0 {
		handler = handler.WithAttrs(opts.Attrs)
	}

	return slog.New(handler), nil
}

func Init(opts Options) (*slog.Logger, error) {
	logger, err := New(opts)
	if err != nil {
		return nil, err
	}

	slog.SetDefault(logger)

	return logger, nil
}

func ParseLevel(name string) (slog.Level, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return slog.LevelInfo, nil
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(name)); err != nil {
		return 0, fmt.Errorf("parse log level %q: %w", name, err)
	}

	return level, nil
}

type ctxKey struct{}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKey{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
			return logger
		}
	}

	return slog.Default()
}
