// Package trace "github.com/actorbuf/iota/trace"
package trace

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// TODO not use Tool
// ToolType
const (
	ToolTraceJaeger        = iota // uber/jaeger-client-go
	ToolTraceOpentelemetry        // opentelemetry
)

var (
	// global tracer
	_tracer    opentracing.Tracer = opentracing.NoopTracer{}
	_driver    Driver
	_config    *Config
	_traceTool int
	_sampler   Sampler
)

// mapCtxSpanKey 改为string类型方便gin的context使用
// 对应 opentracing 的 opentracing.ContextWithSpan
var mapCtxSpanKey string = "iota/trace.span"

func init() {
	// 初始化对应的配置信息为百分百采集
	_config = &Config{
		SamplerConfig: &SamplerConfig{
			New: &SamplerStrategyConfig{
				Type:        SamplerTypeProbability,
				Probability: SamplerProbabilityAlways,
			},
		},
	}
	// 初始化采集器方法
	SetSampler(NewProbabilitySampler)
}

// SetGlobalTracer SetGlobalTracer
func SetGlobalTracer(tracer opentracing.Tracer, traceTool int, driver Driver) {
	_tracer = tracer
	_traceTool = traceTool
	_driver = driver
}

// GetGlobalTracer doSomething
func GetGlobalTracer() opentracing.Tracer {
	return _tracer
}

// GetToolTrace provide tool type
func GetToolTrace() int {
	return _traceTool
}

// SetConfig 设置相关配置
func SetConfig(config *Config) {
	_config = config
	// 再次初始化采集器方法
	SetSampler(NewProbabilitySampler)
}

// SetSampler 设置采集方法
func SetSampler(samplerBuilder SamplerBuilder) {
	if _config == nil || _config.SamplerConfig == nil {
		panic("config is nil")
	}
	_sampler = samplerBuilder(_config.SamplerConfig)
}

// New trace instance with given operationName.
func New(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	child := checkChildFromOpts(opts...)
	// if not ChildOf， use SamplerType
	if !child && !_sampler.OnNew(operationName, opts...).Sample {
		return GetNoopSpan(operationName, opts...)
	}
	// if ChildOf， use SamplerType
	if child && !_sampler.OnNewChildSpan(operationName, opts...).Sample {
		return GetNoopSpan(operationName, opts...)
	}
	return _tracer.StartSpan(operationName, opts...)
}

// checkChildFromOpts check childOf
func checkChildFromOpts(opts ...opentracing.StartSpanOption) bool {
	o := &opentracing.StartSpanOptions{}
	for _, opt := range opts {
		opt.Apply(o)
	}
	for _, reference := range o.References {
		if reference.Type != opentracing.ChildOfRef {
			continue
		}
		return true
	}
	return false
}

// NewWithMapCtx trace instance with given operationName.
func NewWithMapCtx(ctx MapCtx, operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	span := New(operationName, opts...)
	InjectSpanAfterNew(ctx, span)
	return span
}

// Inject takes the Trace instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`.
func Inject(t opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return _tracer.Inject(t, format, carrier)
}

// Extract returns a Trace instance given `format` and `carrier`.
// return `ErrTraceNotFound` if trace not found.
func Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return _tracer.Extract(format, carrier)
}

// NewChildSpan new trace NewChildSpan from Span
func NewChildSpan(childServerName string, spanContext opentracing.SpanContext) opentracing.Span {
	return New(childServerName, opentracing.ChildOf(spanContext))
}

// NewChildSpanWithMapCtx new trace NewChildSpan from Span
// compatibility mapCtx new span; like gin
func NewChildSpanWithMapCtx(ctx MapCtx, childServerName string, spanContext opentracing.SpanContext) opentracing.Span {
	return NewWithMapCtx(ctx, childServerName, opentracing.ChildOf(spanContext))
}

// NewRPCSpan new trace NewRPCSpan from Span
func NewRPCSpan(childServerName string, spanContext opentracing.SpanContext) opentracing.Span {
	return New(childServerName, ext.RPCServerOption(spanContext))
}

// NewRootSpan new rootSpan
func NewRootSpan(rootSpanName string) opentracing.Span {
	return New(rootSpanName)
}

// NewRootSpanWithMapCtx new rootSpan
// compatibility mapCtx new span; like gin
func NewRootSpanWithMapCtx(ctx MapCtx, rootSpanName string) opentracing.Span {
	return NewWithMapCtx(ctx, rootSpanName)
}

// InjectSpanAfterNew Inject span to MapCtx
func InjectSpanAfterNew(ctx MapCtx, span opentracing.Span) {
	IotaSpan := getIotaSpanFromCtx(ctx)
	if IsNoopSpan(IotaSpan) {
		// inject opentracing span
		ctx.Set(mapCtxSpanKey, span)
	}
	if _driver != nil {
		_driver.InjectSpanAfterNew(ctx, span)
	}
	return
}

// getIotaSpanFromCtx get Iota span from ctx
func getIotaSpanFromCtx(ctx context.Context) opentracing.Span {
	// get from iota
	iotaSpan := ctx.Value(mapCtxSpanKey)
	if span, ok := iotaSpan.(opentracing.Span); ok {
		return span
	}
	// get from opentracing
	return opentracing.SpanFromContext(ctx)
}

// NewTracerContext NewContext new a trace context.
// NOTE: This method is not thread safe.
// same as opentracing.ContextWithSpan()
func NewTracerContext(ctx context.Context, span opentracing.Span) context.Context {
	// InjectSpan to MapCtx
	mapCtx, ok := ctx.(MapCtx)
	if ok {
		InjectSpanAfterNew(mapCtx, span)
	}
	// fresh opentracing span ctx
	ctx = opentracing.ContextWithSpan(ctx, span)
	// add opentracing.ctx
	ctx = context.WithValue(ctx, mapCtxSpanKey, span)
	if _driver != nil {
		ctx = _driver.NewTracerContext(ctx, span)
	}
	return ctx
}

// ObtainChildSpan get childSpan from ctx
// if not iota span,return NoopSpan
func ObtainChildSpan(ctx context.Context, name string) opentracing.Span {
	if ctx == nil {
		ctx = context.TODO()
	}
	// find span from ctx
	span := ObtainCtxSpan(ctx)
	if IsNoopSpan(span) {
		return span
	}

	// if has span, new ChildSpan
	mapCtx, ok := ctx.(MapCtx)
	if ok {
		return NewChildSpanWithMapCtx(mapCtx, name, span.Context())
	} else {
		return NewChildSpan(name, span.Context())
	}
}

// ObtainCtxSpan get span from ctx
func ObtainCtxSpan(ctx context.Context) opentracing.Span {
	// find iota span
	iotaSpan := getIotaSpanFromCtx(ctx)
	if iotaSpan != nil {
		return iotaSpan
	}
	return GetNoopSpan("")
}

// ObtainTraceID get TraceID from ctx
func ObtainTraceID(ctx context.Context) string {
	if _driver == nil {
		return ""
	}
	return _driver.ObtainTraceID(ctx)
}

// GetTraceStrFromSpan get TraceStr from Span
func GetTraceStrFromSpan(span opentracing.Span) string {
	if IsNoopSpan(span) {
		return ""
	}
	if _driver == nil {
		return ""
	}
	return _driver.GetTraceStrFromSpan(span)
}

// GetTraceStrFromCtx get TraceStr from ctx
func GetTraceStrFromCtx(ctx context.Context) string {
	spanFace := ObtainCtxSpan(ctx)
	if IsNoopSpan(spanFace) {
		return ""
	}
	if _driver == nil {
		return ""
	}
	return _driver.GetTraceStrFromCtx(ctx)
}

// NewChildSpanFromTraceStr new ChildSpan From TraceStr
// TraceStr:
// 	 - TraceOpentelemetry: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
func NewChildSpanFromTraceStr(childServerName string, traceStr string) (opentracing.Span, error) {
	if _driver == nil {
		return GetNoopSpan(""), nil
	}
	return _driver.NewChildSpanFromTraceStr(childServerName, traceStr)
}

// NewContextFromTraceStr new Ctx From TraceStr
// TraceStr:
// 	 - TraceOpentelemetry: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
func NewContextFromTraceStr(ctx context.Context, childServerName string, traceStr string) (context.Context, error) {
	// new span
	span, err := NewChildSpanFromTraceStr(childServerName, traceStr)
	if err != nil {
		return nil, err
	}
	// finish span, clear span
	span.Finish()
	// if mapCtx Inject span
	mapCtx, ok := ctx.(MapCtx)
	if ok {
		InjectSpanAfterNew(mapCtx, span)
	}
	return context.WithValue(ctx, mapCtxSpanKey, span), nil
}

// MapCtx use set to inject ctx.
// like: gin.Context
type MapCtx interface {
	Set(key string, value interface{})
	context.Context
}
