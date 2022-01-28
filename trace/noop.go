package trace

import (
	"github.com/opentracing/opentracing-go"
)

func IsNoopSpan(span opentracing.Span) bool {
	// is nil or iota NoopSpan ?
	if span == nil {
		return true
	}
	// is opentracing noopSpan ?
	trace := span.Tracer()
	if _, ok := trace.(opentracing.NoopTracer); ok {
		return true
	}
	return false
}

func GetNoopSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return opentracing.NoopTracer{}.StartSpan(operationName, opts...)
}
