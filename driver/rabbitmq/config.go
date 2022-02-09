package rabbitmq

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Address         string                 `yaml:"rabbitmq-address" validate:"required,url"`
	ExchangeName    string                 `yaml:"exchange-name" validate:"required"`
	ExchangeKind    string                 `yaml:"exchange-kind" validate:"required"`
	ExchangeDurable bool                   `yaml:"exchange-durable"`
	QueueName       string                 `yaml:"queue-name" validate:"required"`
	QueueDurable    bool                   `yaml:"queue-durable"`
	BindKey         string                 `yaml:"bind-key" validate:"required"`
	DeliveryMode    uint8                  `yaml:"delivery-mode" validate:"required"` // 2是持久化
	PrefetchCount   int                    `yaml:"prefetch-count" validate:"required,gt=0"`
	PrefetchSize    int                    `yaml:"prefetch-size"`
	AutoAck         bool                   `yaml:"auto-ack"`
	Exclusive       bool                   `yaml:"exclusive"`
	NoLocal         bool                   `yaml:"no-local"`
	NoWait          bool                   `yaml:"no-wait"`
	Args            map[string]interface{} `yaml:"args"`
	ChannelNum      int                    `yaml:"channel-num"` // 单连接多 channel 配置
}

func (c *Config) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	return err
}

type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}
