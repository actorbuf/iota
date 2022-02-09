package rabbitmq

import (
	"context"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RabbitMQ struct {
	ctx      context.Context
	alarm    Alarm
	consumer *consumer
	producer *producer
	handle   func(context.Context, *amqp.Delivery) Action
	config   *Config
	logger   Logger
}

func (r *RabbitMQ) WithContext(ctx context.Context) *RabbitMQ {
	r.ctx = ctx
	return r
}

func (r *RabbitMQ) SetAlarm(alarm Alarm) *RabbitMQ {
	r.alarm = alarm
	return r
}

func (r *RabbitMQ) WithLogger(logger Logger) *RabbitMQ {
	r.logger = logger
	return r
}

func (r *RabbitMQ) SetConfig(config *Config) *RabbitMQ {
	r.config = config
	return r
}

func (r *RabbitMQ) SetHandle(handle func(context.Context, *amqp.Delivery) Action) *RabbitMQ {
	r.handle = handle
	return r
}

func (r *RabbitMQ) Validate() error {
	if r.ctx == nil {
		r.ctx = context.TODO()
	}
	if r.config == nil {
		return errors.New("config is nil")
	}
	return nil
}

func (r *RabbitMQ) StartConsumer() {
	if err := r.Validate(); err != nil {
		panic(err)
	}
	// 初始化队列
	r.consumer = &consumer{
		ctx:   r.ctx,
		alarm: r.alarm,
		channel: &channel{
			config: r.config,
		},
	}
	// 检查是否有 Channel
	if err := r.consumer.Validate(); err != nil {
		panic(err)
		return
	}
	if r.handle == nil {
		panic(errors.New("handle is nil"))
		return
	}
	go r.consumer.Run(r.handle)
	go r.consumer.gracefulShutdown()
}

func (r *RabbitMQ) StartProducer() (chan<- []byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	var err error
	r.producer = &producer{
		key:            r.config.BindKey,
		exchange:       r.config.ExchangeName,
		deliveryMode:   r.config.DeliveryMode,
		sendBodyLength: 4096,
		alarm:          r.alarm,
	}
	r.producer.channel = &channel{
		config: r.config,
	}
	r.config.ChannelNum = 1
	err = r.producer.channel.Init()
	if err != nil {
		return nil, err
	}
	err = r.producer.Validate()
	if err != nil {
		return nil, err
	}

	go r.producer.Run()
	time.Sleep(time.Second)
	return r.producer.sendBody, nil
}
