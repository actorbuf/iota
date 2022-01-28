package otel_jaeger

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"testing"
	"time"
)

func TestOpentelemetryTrace(t *testing.T) {
	defer func() {
		time.Sleep(time.Second * 2)
	}()
	cfg := &Config{
		ServiceName: "iota",
		AgentHost:   "127.0.0.1",
	}
	tp, err := tracerProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("component-main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()
	span.SetAttributes(attribute.Key("main1111").String("value1111"))

	// Use the global TracerProvider.
	tr2 := otel.Tracer("component-bar")
	ctx, span2 := tr2.Start(ctx, "bar")
	span2.SetAttributes(attribute.Key("testset111").String("value1111"))
	defer span2.End()

}

func TestOpentelemetry(t *testing.T) {
	cfg := &Config{
		ServiceName: "iota",
		AgentHost:   "127.0.0.1",
	}
	tp, err := tracerProvider(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
	}(ctx)

	tr := tp.Tracer("component-main")

	ctx, span := tr.Start(ctx, "foo")
	defer span.End()

	bar(ctx)
}

func bar(ctx context.Context) {
	// Use the global TracerProvider.
	tr := otel.Tracer("component-bar")
	ctx, span := tr.Start(ctx, "bar")
	span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()
}
