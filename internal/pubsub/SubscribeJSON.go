package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	// kind ExchangeKind,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType) //kind
	if err != nil {
		return err
	}

	deliveryChannel, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveryChannel {
			var message T
			json.Unmarshal(delivery.Body, &message)
			ackType := handler(message)

			switch ackType {
			case Ack:
				delivery.Ack(false)
				fmt.Println("Message: Acked")
			case NackRequeue:
				delivery.Nack(false, true)
				fmt.Println("Message: NackRequeue")
			case NackDiscard:
				delivery.Nack(false, false)
				fmt.Println("Message: NackDiscard")
			}
		}
	}()

	return nil
}
