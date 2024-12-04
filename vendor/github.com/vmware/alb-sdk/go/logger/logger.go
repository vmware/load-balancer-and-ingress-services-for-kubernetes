package logger

import (
	"context"

	"github.com/vmware/alb-sdk/go/models"
)

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceIDType := ctx.Value(models.TraceID)
	traceID, ok := traceIDType.(string)
	if !ok {
		return ""
	}
	return traceID
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, models.TraceID, traceID)
}
