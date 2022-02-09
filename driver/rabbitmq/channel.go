package rabbitmq

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type channel struct {
	logger Logger
	Chan   []*amqp.Channel
	Conn   *amqp.Connection
	config *Config
	l      sync.Mutex
}

func (r *channel) Close() {
	for i := range r.Chan {
		r.Chan[i].Close()
	}
	r.Conn.Close()
}

func (r *channel) Init() error {
	// 配置检验
	if r.config == nil {
		return errors.New("config is nil")
	}
	if err := r.config.Validate(); err != nil {
		return err
	}

	if r.Conn != nil {
		r.Conn.Close()
	}
	if len(r.Chan) != 0 {
		for i := range r.Chan {
			r.Chan[i].Close()
		}
		r.Chan = make([]*amqp.Channel, 0)
	}
	// 连接初始化
	var err error
	r.Conn, err = amqp.Dial(r.config.Address)
	if err != nil {
		return err
	}
	for i := 0; i < r.config.ChannelNum; i++ {
		newChan, err := r.Conn.Channel()
		if err != nil {
			return err
		}
		err = newChan.Qos(r.config.PrefetchCount, r.config.PrefetchSize, false)
		if err != nil {
			return err
		}
		r.Chan = append(r.Chan, newChan)
	}

	exchange := Exchange{
		Name:    r.config.ExchangeName,
		Kind:    r.config.ExchangeKind,
		Durable: r.config.ExchangeDurable,
	}
	queue := Queue{
		Name:    r.config.QueueName,
		Durable: r.config.QueueDurable,
	}

	if r.config.Args != nil {
		queue.Args = r.config.Args
	}

	// 绑定
	err = r.Bind(exchange, queue, r.config.BindKey, r.config.NoWait, nil)
	if err != nil {
		return err
	}
	return nil
}

func (r *channel) Bind(ex Exchange, qu Queue, key string, noWait bool, args map[string]interface{}) error {
	for i := range r.Chan {
		err := r.Chan[i].ExchangeDeclare(
			ex.Name,
			ex.Kind,
			ex.Durable,
			ex.AutoDelete,
			ex.Internal,
			ex.NoWait,
			ex.Args,
		)
		if err != nil {
			return err
		}
		_, err = r.Chan[i].QueueDeclare(
			qu.Name,
			qu.Durable,
			qu.AutoDelete,
			qu.Exclusive,
			qu.NoWait,
			qu.Args,
		)
		if err != nil {
			return err
		}
		err = r.Chan[i].QueueBind(
			qu.Name, // queue name, 这里指的是 test_logs
			key,     // routing key
			ex.Name, // exchange
			noWait,
			args,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
