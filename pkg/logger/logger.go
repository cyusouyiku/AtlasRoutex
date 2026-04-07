package logger

import "context"

type Logger interface {
Info(ctx context.Context, msg string, kv ...any)
Error(ctx context.Context, msg string, kv ...any)
Debug(ctx context.Context, msg string, kv ...any)
}
