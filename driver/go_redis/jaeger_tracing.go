package goRedis

import (
	"context"
	"github.com/actorbuf/iota/trace"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tracerLog "github.com/opentracing/opentracing-go/log"
	"strconv"
)

const (
	operationRedis  = "Redis-"
	logCmdName      = "command"
	logCmdArgs      = "args"
	logCmdResult    = "result"
	traceTagTraceID = "TraceID"
	traceTagSpanID  = "SpanID"
)

type contextKey int

const (
	cmdStart contextKey = iota
)

// jaegerHook is go-redis jaeger hook
type jaegerHook struct{}

// NewJaegerHook return jaegerHook
func NewJaegerHook() redis.Hook {
	return &jaegerHook{}
}

// BeforeProcess redis before execute action do something
func (jh *jaegerHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span := trace.ObtainChildSpan(ctx, operationRedis+cmd.Name())

	ctx = context.WithValue(ctx, cmdStart, span)
	return ctx, nil
}

// AfterProcess redis after execute action do something
func (jh *jaegerHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span, ok := ctx.Value(cmdStart).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	setCommonTag(ctx, span)

	span.LogFields(tracerLog.String(logCmdName, cmd.Name()))
	span.LogFields(tracerLog.Object(logCmdArgs, cmd.Args()))
	span.LogFields(tracerLog.Object(logCmdResult, cmd.String()))

	if err := cmd.Err(); isRedisError(err) {
		span.LogFields(tracerLog.Error(err))
		span.SetTag(string(ext.Error), true)
	}

	return nil
}

// BeforeProcessPipeline before command process handle
func (jh *jaegerHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	span := trace.ObtainChildSpan(ctx, operationRedis+"pipeline")

	ctx = context.WithValue(ctx, cmdStart, span)

	return ctx, nil
}

// AfterProcessPipeline after command process handle
func (jh *jaegerHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span, ok := ctx.Value(cmdStart).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	hasErr := false
	for idx, cmd := range cmds {
		if err := cmd.Err(); isRedisError(err) {
			hasErr = true
		}
		span.LogFields(tracerLog.String(jh.getPipeLineLogKey(logCmdName, idx), cmd.Name()))
		span.LogFields(tracerLog.Object(jh.getPipeLineLogKey(logCmdArgs, idx), cmd.Args()))
		span.LogFields(tracerLog.String(jh.getPipeLineLogKey(logCmdResult, idx), cmd.String()))
	}
	if !hasErr {
		span.SetTag(string(ext.Error), true)
	}

	return nil
}

func (jh *jaegerHook) getPipeLineLogKey(logField string, idx int) string {
	return logField + "-" + strconv.Itoa(idx)
}

func isRedisError(err error) bool {
	if err == redis.Nil {
		return false
	}
	_, ok := err.(redis.Error)
	return ok
}

// setCommonTag 设置公众的traceTag
func setCommonTag(ctx context.Context, span opentracing.Span) {
	// 植入traceID 与 spanID
	if traceID := trace.ObtainTraceID(ctx); traceID != "" {
		span.SetTag(traceTagTraceID, traceID)
	}
	if spanID := trace.ObtainSpanID(ctx); spanID != "" {
		span.SetTag(traceTagSpanID, spanID)
	}
}
