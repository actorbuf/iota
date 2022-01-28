package trace

import (
	"context"
	"github.com/opentracing/opentracing-go"
)

type Driver interface {
	// InjectSpanAfterNew 新建span后注入span到 MapCtx
	InjectSpanAfterNew(ctx MapCtx, span opentracing.Span)
	// NewTracerContext 返回带 span 的ctx （用于注入三方ctx中的span）
	NewTracerContext(ctx context.Context, span opentracing.Span) context.Context
	// ObtainTraceID ctx 中获取 traceID
	ObtainTraceID(ctx context.Context) string
	// GetTraceStrFromSpan span中获取 traceStr
	GetTraceStrFromSpan(span opentracing.Span) string
	// GetTraceStrFromCtx ctx 中获取 traceStr
	GetTraceStrFromCtx(ctx context.Context) string
	// NewChildSpanFromTraceStr 根据 traceStr 创建span
	NewChildSpanFromTraceStr(childServerName string, traceStr string) (opentracing.Span, error)
}
