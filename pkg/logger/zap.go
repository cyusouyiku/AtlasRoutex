package logger

import (
"context"
"log/slog"
"os"
)

type SLogger struct {
inner *slog.Logger
}

func NewSLogger() *SLogger {
return &SLogger{inner: slog.New(slog.NewTextHandler(os.Stdout, nil))}
}

func (l *SLogger) Info(_ context.Context, msg string, kv ...any)  { l.inner.Info(msg, kv...) }
func (l *SLogger) Error(_ context.Context, msg string, kv ...any) { l.inner.Error(msg, kv...) }
func (l *SLogger) Debug(_ context.Context, msg string, kv ...any) { l.inner.Debug(msg, kv...) }
