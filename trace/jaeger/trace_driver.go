package jaeger

import (
	"context"
	"github.com/actorbuf/iota/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

type UberDriver struct{}

var DefaultUberDriver trace.Driver = (*UberDriver)(nil)

// InjectSpanAfterNew 新建span后注入span到 MapCtx
func (d *UberDriver) InjectSpanAfterNew(_ trace.MapCtx, _ opentracing.Span) {
	return
}

// NewTracerContext 返回带 span 的ctx （用于注入三方ctx中的span）
func (d *UberDriver) NewTracerContext(ctx context.Context, _ opentracing.Span) context.Context {
	return ctx
}

// ObtainTraceID ctx 中获取 traceID
func (d *UberDriver) ObtainTraceID(ctx context.Context) string {
	spanFace := trace.ObtainCtxSpan(ctx)
	if trace.IsNoopSpan(spanFace) {
		return ""
	}
	uSpan, ok := spanFace.Context().(jaeger.SpanContext)
	if !ok {
		return ""
	}
	return uSpan.TraceID().String()
}

// GetTraceStrFromSpan span中获取 traceStr
func (d *UberDriver) GetTraceStrFromSpan(span opentracing.Span) string {
	spCtx, ok := span.Context().(jaeger.SpanContext)
	if !ok {
		return ""
	}
	return spCtx.String()
}

// GetTraceStrFromCtx ctx 中获取 traceStr
func (d *UberDriver) GetTraceStrFromCtx(ctx context.Context) string {
	spanFace := trace.ObtainCtxSpan(ctx)
	if trace.IsNoopSpan(spanFace) {
		return ""
	}
	uSpan, ok := spanFace.Context().(jaeger.SpanContext)
	if !ok {
		return ""
	}
	return uSpan.String()
}

// NewChildSpanFromTraceStr 根据 traceStr 创建span
func (d *UberDriver) NewChildSpanFromTraceStr(childServerName string, traceStr string) (opentracing.Span, error) {
	spCtx, err := jaeger.ContextFromString(traceStr)
	if err != nil {
		return trace.GetNoopSpan(""), err
	}
	return trace.NewChildSpan(childServerName, spCtx), nil
}
