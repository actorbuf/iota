package otel_jaeger

import "github.com/actorbuf/iota/trace"

type Config struct {
	// ServiceName specifies the service name to use on the tracer.
	ServiceName string `yaml:"jaeger-service-name"`
	// AgentHost report trace SamplerType config
	// AgentPort instructs reporter to send spans to jaeger-agent at this address
	AgentHost string `yaml:"jaeger-agent-host"`
	AgentPort string `yaml:"jaeger-agent-port"`

	// Disabled
	Disabled bool `yaml:"jaeger-disabled"`

	// SamplerConfig
	SamplerConfig *trace.SamplerConfig `yaml:"sampler_config"`
}
