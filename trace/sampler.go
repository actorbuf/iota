package trace

import (
	"context"
	"github.com/opentracing/opentracing-go"
)


type SamplerType int32

const (
	// SamplerTypeProbability 随机比例采集
	SamplerTypeProbability SamplerType = iota
)

// SamplerProbabilityAlways 全部采集
const SamplerProbabilityAlways = 1

// SamplerProbabilityFalse 全都不采集
const SamplerProbabilityFalse = 0

type SamplerConfig struct {
	New      *SamplerStrategyConfig `yaml:"new"`
	NewChild *SamplerStrategyConfig `yaml:"new_child"`
}

type SamplerStrategyConfig struct {
	Type        SamplerType `yaml:"type"`
	Probability float64     `yaml:"probability"`
}

// SamplingDecision 采样参数，方便做扩充
type SamplingDecision struct {
	Sample    bool
	Retryable bool
}

// Sampler 采样操作
type Sampler interface {
	// OnNew use when New
	OnNew(operationName string, opts ...opentracing.StartSpanOption) SamplingDecision

	// OnNewChildSpan use when New and build child span
	OnNewChildSpan(operationName string, opts ...opentracing.StartSpanOption) SamplingDecision

	// OnInject use when Inject
	OnInject(t opentracing.SpanContext, format interface{}, carrier interface{}) SamplingDecision

	// OnExtract use when Extract
	OnExtract(format interface{}, carrier interface{}) SamplingDecision

	// OnNewChildSpanFromTraceStr use when NewChildSpanFromTraceStr
	OnNewChildSpanFromTraceStr(childServerName string, traceStr string) SamplingDecision

	// OnNewContextFromTraceStr use when NewContextFromTraceStr
	OnNewContextFromTraceStr(ctx context.Context, childServerName string, traceStr string) SamplingDecision

	// Close if need close Sampler
	Close()
}

// SamplerBuilder Sampler 创建者
type SamplerBuilder = func(samplerConfig *SamplerConfig) Sampler

var _ Sampler = (*ProbabilitySampler)(nil)

// ProbabilitySampler 按概率进行采样
type ProbabilitySampler struct {
	SamplerConfig *SamplerConfig
}

// OnNew 按概率看是否创建新的span
func (p *ProbabilitySampler) OnNew(operationName string, opts ...opentracing.StartSpanOption) SamplingDecision {
	samplingDecision := SamplingDecision{}
	if p.SamplerConfig == nil || p.SamplerConfig.New == nil {
		return samplingDecision

	}
	switch p.SamplerConfig.New.Type {
	case SamplerTypeProbability:
		// 百分百
		if p.SamplerConfig.New.Probability == SamplerProbabilityAlways {
			samplingDecision.Sample = true
			return samplingDecision
		}
		// 一定不做
		if p.SamplerConfig.New.Probability == SamplerProbabilityFalse {
			samplingDecision.Sample = false
			return samplingDecision
		}
		// 放大一百
		intD := int64(p.SamplerConfig.New.Probability * 100)
		randD := RandInt64(0, 100)
		samplingDecision.Sample = randD < intD
		return samplingDecision
	default:
		return samplingDecision
	}
}

func (p *ProbabilitySampler) OnNewChildSpan(operationName string, opts ...opentracing.StartSpanOption) SamplingDecision {
	return SamplingDecision{Sample: true, Retryable: true}
}

func (p *ProbabilitySampler) OnInject(t opentracing.SpanContext, format interface{}, carrier interface{}) SamplingDecision {
	return SamplingDecision{Sample: true, Retryable: true}
}

func (p *ProbabilitySampler) OnExtract(format interface{}, carrier interface{}) SamplingDecision {
	return SamplingDecision{Sample: true, Retryable: true}
}

func (p *ProbabilitySampler) OnNewChildSpanFromTraceStr(childServerName string, traceStr string) SamplingDecision {
	return SamplingDecision{Sample: true, Retryable: true}
}

func (p *ProbabilitySampler) OnNewContextFromTraceStr(ctx context.Context, childServerName string, traceStr string) SamplingDecision {
	return SamplingDecision{Sample: true, Retryable: true}
}

func (p *ProbabilitySampler) Close() {
	return
}

// NewProbabilitySampler 创建概率采集器
func NewProbabilitySampler(samplerConfig *SamplerConfig) Sampler {
	return &ProbabilitySampler{SamplerConfig: samplerConfig}
}
