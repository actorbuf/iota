package rabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	new(RabbitMQ).SetConfig(&Config{
		Address:         "amqp://admin:admin@10.0.0.135:5672/default",
		ExchangeName:    "mq_test",
		ExchangeKind:    "topic",
		ExchangeDurable: true,
		QueueName:       "mq_test",
		QueueDurable:    true,
		BindKey:         "mq_test",
		DeliveryMode:    2,
		PrefetchCount:   5,
	}).
		SetHandle(func(ctx context.Context, delivery *amqp.Delivery) Action {
			for i := 0; i < 10; i++ {
				fmt.Println(i)
				time.Sleep(time.Second)
			}
			fmt.Println(string(delivery.Body))
			return Ack
		}).
		StartConsumer()
	ch1 := make(chan os.Signal)
	signal.Notify(ch1, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)
		_ = <-ch
		fmt.Println("1")
		time.Sleep(2 * time.Second)
		fmt.Println("2")
	}()
	<-ch1
	fmt.Println("quit")
}
