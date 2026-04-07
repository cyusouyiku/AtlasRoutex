package interceptor

import (
"context"
"time"
)

func Logging(ctx context.Context, begin time.Time) (context.Context, time.Duration) {
return ctx, time.Since(begin)
}
