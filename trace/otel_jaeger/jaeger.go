package otel_jaeger // Package otel_jaeger "github.com/actorbuf/iota/trace/otel_jaeger"

import (
	"context"
	"fmt"
	"github.com/actorbuf/iota/trace"
	"github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/otel"
	otelo "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"io"
	"log"
	"os"
	"time"
)

const service = "iota"

func initIotaTraceConfig(config *Config) *trace.Config {
	iotaTraceConfig := &trace.Config{
		SamplerConfig: config.SamplerConfig,
	}
	return iotaTraceConfig
}

type NullCloser struct{}

func (*NullCloser) Close() error { return nil }

type JaegerCloser struct {
	shutdown func() error
}

func (c *JaegerCloser) Close() error { return c.shutdown() }

func NewJaeger(config *Config) (io.Closer, error) {
	tp, err := tracerProvider(config)
	if err != nil {
		return &trace.NullCloser{}, err
	}
	// disable return nil
	if tp == nil {
		return &trace.NullCloser{}, nil
	}
	closer := &JaegerCloser{
		shutdown: func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return tp.Shutdown(ctx)
		},
	}

	// TODO the bridge emits warnings because the Jaeger exporter does not defer context setup
	_, bridgeTracer, wrapperProvider := otelo.NewTracerPairWithContext(context.Background(), tp.Tracer("OpenTracing"))
	bridgeTracer.SetWarningHandler(func(msg string) {
		// TODO inject warning handler
		fmt.Printf("bridgeTracer warning %s \n", msg)
	})
	propagator := propagation.NewCompositeTextMapPropagator(propagation.Baggage{}, propagation.TraceContext{})
	bridgeTracer.SetTextMapPropagator(propagator)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(wrapperProvider)
	otel.SetTextMapPropagator(propagator)
	otel.SetErrorHandler(otelErrorHandlerFunc(func(err error) {
		fmt.Println("msg", "OpenTelemetry.ErrorHandler", "err", err)
	}))

	// set otel bridgeTracer
	opentracing.SetGlobalTracer(bridgeTracer)

	// set iota globalTracer
	trace.SetGlobalTracer(opentracing.GlobalTracer(), trace.ToolTraceOpentelemetry, DefaultOpentelemetryDriver)
	trace.SetConfig(initIotaTraceConfig(config))
	return closer, nil
}

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(config *Config) (*tracesdk.TracerProvider, error) {
	// Disabled return nil
	if config.Disabled {
		return nil, nil
	}
	// Create the Jaeger exporter
	exp, err := jaeger.New(
		jaeger.WithAgentEndpoint(jaeger.WithAgentHost(config.AgentHost), jaeger.WithAgentPort(config.AgentPort), jaeger.WithLogger(log.New(os.Stdout, "", 0))),
	)

	//exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp, tracesdk.WithMaxQueueSize(200)),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)),
	)
	return tp, nil
}

type otelErrorHandlerFunc func(error)

// Handle implements otel.ErrorHandler
func (f otelErrorHandlerFunc) Handle(err error) {
	f(err)
}
