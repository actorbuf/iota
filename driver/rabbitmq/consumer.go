package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type consumer struct {
	ctx     context.Context
	logger  Logger
	alarm   Alarm
	channel *channel
	close   bool
	wait    *sync.WaitGroup
}

func (c *consumer) Validate() error {
	if c.channel == nil {
		return errors.New("channel is nil")
	}
	if c.logger == nil {
		c.logger = NewDefaultLogger()
	}
	if c.channel.config == nil {
		return errors.New("channel config is nil")
	}
	if c.channel.config.ChannelNum == 0 {
		c.channel.config.ChannelNum = 0
	}
	if c.channel.config.QueueName == "" {
		return errors.New("queue is nil")
	}
	if c.wait == nil {
		c.wait = &sync.WaitGroup{}
	}
	return nil
}

func (c *consumer) GetDelivery(index int) (<-chan amqp.Delivery, error) {
	if len(c.channel.Chan)-1 < index {
		return nil, errors.New("out of range")
	}
	delivery, err := c.channel.Chan[index].Consume(
		c.channel.config.QueueName, // queue
		"",                         // consumer
		c.channel.config.AutoAck,   // auto-ack
		c.channel.config.Exclusive, // exclusive
		c.channel.config.NoLocal,   // no-local
		c.channel.config.NoWait,    // no-wait
		c.channel.config.Args,      // args
	)
	if err != nil {
		// 失败重新连接MQ
		return nil, errors.New("get deliver error, " + err.Error())
	}
	return delivery, nil
}

func (c *consumer) GetConn() error {
	c.channel.l.Lock()
	defer c.channel.l.Unlock()
	if c.channel.Conn != nil {
		if !c.channel.Conn.IsClosed() {
			return nil
		}
	}
	var reconnectCount = 0
	var maxReconnectCount = 3
	var alarmFlag bool
	for {
		time.Sleep(time.Duration(reconnectCount*reconnectCount) * time.Second)
		err := c.channel.Init()
		if err != nil {
			if reconnectCount >= maxReconnectCount {
				if !alarmFlag {
					fmt.Printf("RabbitMQ-Consumer err: %+v", err)
					if c.alarm != nil {
						_ = c.alarm.SetMsg(map[string]string{
							"Title":   "RabbitMQ 连接失败超出阈值",
							"Address": c.channel.config.Address,
							"Queue":   c.channel.config.QueueName,
						})
						_ = c.alarm.Do()
					}
					alarmFlag = true
				}
			}
			// 指数退让重试
			reconnectCount++
			continue
		}
		return nil
	}
}

type Action int

const (
	// Ack default ack this msg after you have successfully processed this delivery.
	Ack Action = iota
	// NackDiscard the message will be dropped or delivered to a server configured dead-letter queue.
	NackDiscard
	// NackRequeue deliver this message to a different consumer.
	NackRequeue
)

func (c *consumer) Run(handler func(context.Context, *amqp.Delivery) Action) {
RECONNECT:
	err := c.GetConn()
	if err != nil {
		time.Sleep(3 * time.Second)
		goto RECONNECT
	}

	for i := range c.channel.Chan {
		go c.doHandlerLoop(c.ctx, i, handler)
	}
}

func (c *consumer) reCreateChannel(index int) error {
	// 判断连接是否被关闭了
	if c.channel.Conn.IsClosed() {
		err := c.GetConn()
		if err != nil {
			return err
		}
	}
	if len(c.channel.Chan)-1 < index {
		return errors.New("out of range")
	}
	c.channel.Chan[index].Close()
	newChan, err := c.channel.Conn.Channel()
	if err != nil {
		return err
	}
	err = newChan.Qos(c.channel.config.PrefetchCount, c.channel.config.PrefetchSize, false)
	if err != nil {
		return err
	}
	c.channel.Chan[index] = newChan
	return nil
}

func (c *consumer) doHandlerLoop(ctx context.Context, channelIndex int, handler func(context.Context, *amqp.Delivery) Action) {
RUNLOOP:
	cc := make(chan *amqp.Error)
	c.channel.Chan[channelIndex].NotifyClose(cc)
	delivery, err := c.GetDelivery(channelIndex)
	if err != nil {
		c.logger.Error("consumer.doHandlerLoop", zap.String("info", fmt.Sprintf("can't get delivery: %+v", err)))
		time.Sleep(3 * time.Second)
		goto RUNLOOP
	}
	for {
		select {
		case msg, ok := <-delivery:
			if !ok {
				if !c.close {
					c.reCreateChannel(channelIndex)
					goto RUNLOOP
				} else {
					return
				}
			}
			if !c.close {
				c.wait.Add(1)
				switch handler(ctx, &msg) {
				case Ack:
					err := msg.Ack(false)
					if err != nil {
						c.logger.Error("consumer.doHandlerLoop", zap.String("info", "can't ack message: "+err.Error()))
					}
				case NackDiscard:
					err := msg.Nack(false, false)
					if err != nil {
						c.logger.Error("consumer.doHandlerLoop", zap.String("info", "can't nack message: "+err.Error()))
					}
				case NackRequeue:
					err := msg.Nack(false, true)
					if err != nil {
						c.logger.Error("consumer.doHandlerLoop", zap.String("info", "can't nack message: "+err.Error()))
					}
				}
				c.wait.Done()
			} else {
				return
			}

		case <-cc:
			if !c.close {
				c.reCreateChannel(channelIndex)
				goto RUNLOOP
			} else {
				return
			}
		}
	}
}

func (c *consumer) gracefulShutdown() {
	// 阻塞，直到接收到shutdown的信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)
	_ = <-ch
	//关闭后，Run方法中处理消息的协程将会关闭，不再处理新消息。
	c.close = true
	c.wait.Wait()
}
