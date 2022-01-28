package test

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go/log"
	"github.com/actorbuf/iota/trace"
	"github.com/actorbuf/iota/trace/otel_jaeger"
	"testing"
	"time"
)

func TestOpentelemetryBridge(t *testing.T) {
	defer func() {
		time.Sleep(time.Second * 6)
	}()

	cfg := &otel_jaeger.Config{
		ServiceName: "iota",
		AgentHost:   "127.0.0.1",
	}
	closer, err := otel_jaeger.NewJaeger(cfg)
	if err != nil {
		panic(err)
	}
	defer func() { _ = closer.Close() }()

	span1 := trace.NewRootSpan("span1")
	span1.LogKV("hello111", "leihou111")
	span1.SetTag("helloTag", "leihouTag")
	span1.LogFields(
		log.String("hello.req", "req"),
		log.String("hello.resp", "resp"),
	)
	defer span1.Finish()

	ctx := context.Background()
	ctx = trace.NewTracerContext(ctx, span1)

	traceStr1 := trace.GetTraceStrFromCtx(ctx)
	fmt.Println("traceStr1", traceStr1)

	// other service
	span2 := trace.ObtainChildSpan(ctx, "span2")
	span2.LogKV("span2", "span2")
	span2.SetTag("span2Tag", "span2Tag")
	span2.LogFields(
		log.String("span2.req", "req"),
		log.String("span2.resp", "resp"),
	)
	defer span2.Finish()

	traceStr := trace.GetTraceStrFromSpan(span2)
	fmt.Println("traceStr", traceStr)

	span3, err := trace.NewChildSpanFromTraceStr("span3", traceStr)
	if err != nil {
		panic(err)
	}
	span3.LogKV("span3", "span3")
	span3.SetTag("span3Tag", "span3Tag")
	span3.LogFields(
		log.String("span3.req", "req"),
		log.String("span3.resp", "resp"),
	)
	defer span3.Finish()

	ctx2, err := trace.NewContextFromTraceStr(context.Background(), "span4", traceStr)
	if err != nil {
		panic(err)
	}
	span4 := trace.ObtainCtxSpan(ctx2)
	span4.LogKV("span4", "span4")
	span4.SetTag("span4Tag", "span4Tag")
	span4.LogFields(
		log.String("span4.req", "req"),
		log.String("span4.resp", "resp"),
	)
	defer span4.Finish()

	span5, err := trace.NewChildSpanFromTraceStr("span5", traceStr1)
	if err != nil {
		panic(err)
	}
	span5.LogKV("span5", "span5")
	span5.SetTag("span5Tag", "span5Tag")
	span5.LogFields(
		log.String("span5.req", "req"),
		log.String("span5.resp", "resp"),
	)
	defer span5.Finish()

	ctx3, err := trace.NewContextFromTraceStr(context.Background(), "span6", traceStr1)
	if err != nil {
		panic(err)
	}
	span6 := trace.ObtainCtxSpan(ctx3)
	span6.LogKV("span6", "span6")
	span6.SetTag("span6Tag", "span6Tag")
	span6.LogFields(
		log.String("span6.req", "req"),
		log.String("span6.resp", "resp"),
	)
	defer span6.Finish()
}
