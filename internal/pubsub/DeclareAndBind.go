package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}
	isTransient := simpleQueueType == 1
	queue, err := channel.QueueDeclare(queueName, !isTransient, isTransient, isTransient, false, nil)
	if err != nil {
		log.Fatalf("Could not declare queue: %v", err)
	}

	bindError := channel.QueueBind(queueName, key, exchange, false, nil)
	if bindError != nil {
		log.Fatalf("Could not bind queue: %v", err)
	}

	return channel, queue, nil
}
