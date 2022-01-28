package jaeger

import (
	"context"
	"errors"
	"io"
)

const (
	// TypeConst - for "const" sampler, 0 or 1 for always false/true respectively
	TypeConst SamplerType = "const"
	// TypeProbabilistic - for "probabilistic" sampler, a probability between 0 and 1
	TypeProbabilistic SamplerType = "probabilistic"
	// TypeRateLimiting - for "rateLimiting" sampler, the number of spans per second
	TypeRateLimiting SamplerType = "rateLimiting"
)

const (
	DefaultConstParam         SamplerParam = 1
	DefaultProbabilisticParam SamplerParam = 0.001
	DefaultRateLimitingParam  SamplerParam = 1000
)

type SamplerType string

type SamplerParam float64

var jaegerCloser io.Closer

func GetJaegerCloser() io.Closer {
	return jaegerCloser
}

type Config struct {
	// ServiceName specifies the service name to use on the tracer.
	ServiceName string `yaml:"jaeger-service-name"`
	// report trace Sampler config
	SamplerCfg SamplerCfg `yaml:"jaeger-sampler-cfg"`
	//AgentHostPort instructs reporter to send spans to jaeger-agent at this address
	AgentHost string `yaml:"jaeger-agent-host"`
	AgentPort string `yaml:"jaeger-agent-port"`
	// LogSpans, when true, enables LoggingReporter that runs in parallel with the main reporter
	// and logs all submitted spans. Main Configuration.Logger must be initialized in the code
	// for this option to have any effect.
	LogSpans bool `yaml:"jaeger-log-spans"`

	Disabled bool `yaml:"jaeger-disabled"`
}

func (cfg *Config) Init(_ context.Context) error {
	if cfg == nil {
		return nil
	}

	var err error
	jaegerCloser, err = NewJaeger(cfg)
	return err
}

type SamplerCfg struct {
	// Type specifies the type of the sampler: const, probabilistic, rateLimiting, or remote
	Type SamplerType `yaml:"jaeger-sampler-type"`
	// Param is a value passed to the sampler.
	// Valid values for Param field are:
	// - for "const" sampler, 0 or 1 for always false/true respectively
	// - for "probabilistic" sampler, a probability between 0 and 1
	// - for "rateLimiting" sampler, the number of spans per second
	// - for "remote" sampler, param is the same as for "probabilistic"
	//   and indicates the initial sampling rate before the actual one
	//   is received from the mothership.
	Param SamplerParam `yaml:"jaeger-sampler-param"`
}

//Validate Config
func (cfg *Config) Validate() error {
	if cfg.SamplerCfg.Type == "" {
		cfg.SamplerCfg.Type = TypeProbabilistic
	}
	switch cfg.SamplerCfg.Type {
	case TypeConst:
		if !(cfg.SamplerCfg.Param == 0.0 || cfg.SamplerCfg.Param == 1.0) {
			cfg.SamplerCfg.Param = DefaultConstParam
		}
	case TypeProbabilistic:
		if !(cfg.SamplerCfg.Param > 0.0 && cfg.SamplerCfg.Param < 1.0) {
			cfg.SamplerCfg.Param = DefaultProbabilisticParam
		}
	case TypeRateLimiting:
		if !(cfg.SamplerCfg.Param > 1) {
			cfg.SamplerCfg.Param = DefaultRateLimitingParam
		}
	default:
		return errors.New("type is not allow")
	}
	if cfg.ServiceName == "" {
		return errors.New("serviceName is nil")
	}
	if cfg.AgentHost == "" {
		return errors.New("AgentHost is nil")
	}
	if cfg.AgentPort == "" {
		return errors.New("AgentPort is nil")
	}
	return nil
}
