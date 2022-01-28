package jaeger

import (
	"fmt"
	"github.com/actorbuf/iota/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

func initIotaTraceConfig(config *Config) *trace.Config {
	iotaTraceConfig := &trace.Config{
		SamplerConfig: &trace.SamplerConfig{
			New: &trace.SamplerStrategyConfig{
				Type: trace.SamplerTypeProbability,
			},
		},
	}
	switch config.SamplerCfg.Type {
	case TypeConst:
		if config.SamplerCfg.Param == 1 {
			iotaTraceConfig.SamplerConfig.New.Probability = trace.SamplerProbabilityAlways
		} else {
			iotaTraceConfig.SamplerConfig.New.Probability = trace.SamplerProbabilityFalse
		}
	case TypeProbabilistic:
		iotaTraceConfig.SamplerConfig.New.Probability = float64(config.SamplerCfg.Param)
	default:
		iotaTraceConfig.SamplerConfig.New.Probability = trace.SamplerProbabilityAlways
	}
	return iotaTraceConfig
}

// NewJaeger new a jaeger trace
// Init then defer close
func NewJaeger(config *Config) (io.Closer, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	tracer, closer, err := initJaeger(config)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	// 预先设置Iota采集器配置
	trace.SetConfig(initIotaTraceConfig(config))
	trace.SetGlobalTracer(tracer, trace.ToolTraceJaeger, DefaultUberDriver)

	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

// initJaeger set jaeger for trace
func initJaeger(Config *Config) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  string(TypeConst),
			Param: float64(DefaultConstParam),
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           Config.LogSpans,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", Config.AgentHost, Config.AgentPort),
		},
		ServiceName: Config.ServiceName,
		Disabled:    Config.Disabled,
	}
	return cfg.NewTracer(config.Logger(jaeger.StdLogger))
}
