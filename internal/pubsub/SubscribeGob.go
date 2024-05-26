package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
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
			buffer := bytes.NewBuffer(delivery.Body)
			decoder := gob.NewDecoder(buffer)

			var message T
			err := decoder.Decode(&message)
			if err != nil {
				delivery.Nack(false, false)
				fmt.Println("Message: NackDiscard")
			}

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
