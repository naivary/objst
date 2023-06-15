package logger

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

type Logger struct {
	sl  *slog.Logger
	ctx context.Context
}

func New(ctx context.Context) *Logger {
	return &Logger{
		sl:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
		ctx: ctx,
	}
}

func (l Logger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.sl.ErrorCtx(l.ctx, msg)
}

func (l Logger) Warningf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.sl.WarnCtx(l.ctx, msg)
}

func (l Logger) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.sl.InfoCtx(l.ctx, msg)
}

func (l Logger) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.sl.DebugCtx(l.ctx, msg)
}
