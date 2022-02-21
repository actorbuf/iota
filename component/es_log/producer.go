package es_log

import (
	"context"
	"encoding/json"
	"github.com/actorbuf/iota/driver/rabbitmq"
	"sync"
	"time"
)

var ExLogProducer chan<- []byte
var once sync.Once

// 内部方法
func send(msg []byte) error {
	ExLogProducer <- msg
	return nil
}

func SendLogDataToMq(msg []byte) error {
	return send(msg)
}

func LogMqPush(log *log) error {
	// 启动拓客的生产者
	initProducer(log)
	dataByte, _ := json.Marshal(log.MqData)
	// 推到队列里
	return SendLogDataToMq(dataByte)
}

func initProducer(log *log) {
	once.Do(func() {
		initTime := 1
	initTry:
		// 有配置用配的，没有用默认的
		address := log.MaAddress
		if address == "" {
			// TODO ADD address
			address = ""
		}
		mq := new(rabbitmq.RabbitMQ).SetConfig(&rabbitmq.Config{
			Address:         address,
			ExchangeName:    "heywoods-es-log-exchange",
			ExchangeKind:    "direct",
			ExchangeDurable: true,
			QueueName:       "heywoods-es-log-queue",
			QueueDurable:    true,
			BindKey:         "heywoods-es-log-routingkey",
			DeliveryMode:    2,
		})
		if log.MqLogger != nil {
			mq.WithLogger(log.MqLogger)
		}
		var err error

		ExLogProducer, err = mq.StartProducer()
		if err != nil {
			log.alarmError(context.Background(), "initProducer", err)
			log.error(context.Background(), "initProducer", err)
			return
		}
		if err != nil {
			// 重试三次
			if initTime > 3 {
				log.alarmError(context.Background(), "initProducer", err)
				log.error(context.Background(), "initProducer", err)
			} else {
				initTime++
				time.Sleep(time.Second * time.Duration(3))
				goto initTry
			}
		} else {
			log.info(context.Background(), "initProducer", "ExLogProducer init success")
		}
	})
}
