package otel_jaeger

import (
	"context"
	"errors"
	"fmt"
	"github.com/actorbuf/iota/trace"
	"github.com/opentracing/opentracing-go"
	otelo "go.opentelemetry.io/otel/bridge/opentracing"
	otelt "go.opentelemetry.io/otel/trace"
	"strings"
)

type OpentelemetryDriver struct{}

var DefaultOpentelemetryDriver trace.Driver = (*OpentelemetryDriver)(nil)

// ctxKeyOtelSpan 改为string类型方便Set类型的context传递（如：gin）
// 对应 opentelemetry 的 otelr.ContextWithSpan
const ctxKeyOtelSpan = "iota/trace.span.otel"

// getOtelSpanFromCtx get otel span from ctx
func getOtelSpanFromCtx(ctx context.Context) otelt.Span {
	otelSpan := ctx.Value(ctxKeyOtelSpan)
	if otelSpan != nil {
		span, _ := otelSpan.(otelt.Span)
		return span
	}
	span := otelt.SpanFromContext(ctx)
	if span.IsRecording() {
		return span
	}
	return nil
}

var ErrTraceFormat = errors.New("trace format error")
var ErrBridgeTracerNotFount = errors.New("bridge tracer not found")

// newChildSpanFromTraceStr
// traceStr format: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
func newChildSpanFromTraceStr(childServerName, traceStr string) (opentracing.Span, error) {
	ta := strings.Split(traceStr, "-")
	if len(ta) != 4 {
		return nil, ErrTraceFormat
	}
	traceID, err := otelt.TraceIDFromHex(ta[1])
	if err != nil {
		return nil, err
	}

	spanID, err := otelt.SpanIDFromHex(ta[2])
	if err != nil {
		return nil, err
	}

	// restart parent span
	spCtx := otelt.NewSpanContext(otelt.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: otelt.FlagsSampled,
		Remote:     true,
	})

	ctx := otelt.ContextWithSpanContext(context.Background(), spCtx)
	parentSpan := otelt.SpanFromContext(ctx)
	// end parentSpan
	defer parentSpan.End()

	ot := opentracing.GlobalTracer()
	bt, ok := ot.(*otelo.BridgeTracer)
	if !ok {
		return nil, ErrBridgeTracerNotFount
	}
	ctx = bt.ContextWithBridgeSpan(ctx, parentSpan)
	span, _ := opentracing.StartSpanFromContext(ctx, childServerName)
	return span, nil
}

// InjectSpanAfterNew 新建span后注入span到 MapCtx
func (d *OpentelemetryDriver) InjectSpanAfterNew(ctx trace.MapCtx, span opentracing.Span) {
	otSpan := getOtelSpanFromCtx(ctx)
	if otSpan == nil {
		// inject otel Span
		oCtx := opentracing.ContextWithSpan(ctx, span)
		otSpan = otelt.SpanFromContext(oCtx)
		if otSpan.IsRecording() {
			ctx.Set(ctxKeyOtelSpan, otSpan)
		}
	}
}

// NewTracerContext 返回带 span 的ctx （用于注入三方ctx中的span）
func (d *OpentelemetryDriver) NewTracerContext(ctx context.Context, _ opentracing.Span) context.Context {
	// add Opentelemetry.ctx
	// inject otel Span
	otelSpan := getOtelSpanFromCtx(ctx)
	if otelSpan != nil {
		ctx = context.WithValue(ctx, ctxKeyOtelSpan, otelSpan)
	}
	return ctx
}

// ObtainTraceID ctx 中获取 traceID
func (d *OpentelemetryDriver) ObtainTraceID(ctx context.Context) string {
	otelSpan := getOtelSpanFromCtx(ctx)
	if otelSpan != nil && otelSpan.IsRecording() {
		return otelSpan.SpanContext().TraceID().String()
	}
	return ""
}

// GetTraceStrFromSpan span中获取 traceStr
func (d *OpentelemetryDriver) GetTraceStrFromSpan(span opentracing.Span) string {
	// use opentracing.ContextWithSpan to set opentracing.span and ot.Span into ctx.
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	otSpan := otelt.SpanFromContext(ctx)
	if !otSpan.IsRecording() {
		return ""
	}
	return fmt.Sprintf("00-%s-%s-%s", otSpan.SpanContext().TraceID().String(), otSpan.SpanContext().SpanID().String(), otSpan.SpanContext().TraceFlags().String())
}

// GetTraceStrFromCtx ctx 中获取 traceStr
func (d *OpentelemetryDriver) GetTraceStrFromCtx(ctx context.Context) string {
	otelSpan := getOtelSpanFromCtx(ctx)
	if otelSpan == nil || !otelSpan.IsRecording() {
		return ""
	}
	// TODO what version, now default use 00
	return fmt.Sprintf("00-%s-%s-%s", otelSpan.SpanContext().TraceID().String(), otelSpan.SpanContext().SpanID().String(), otelSpan.SpanContext().TraceFlags().String())
}

// NewChildSpanFromTraceStr 根据 traceStr 创建span
func (d *OpentelemetryDriver) NewChildSpanFromTraceStr(childServerName string, traceStr string) (opentracing.Span, error) {
	span, err := newChildSpanFromTraceStr(childServerName, traceStr)
	if err != nil || span == nil {
		return trace.GetNoopSpan(""), err
	}
	return span, nil
}
