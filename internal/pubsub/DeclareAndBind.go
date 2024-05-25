package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable   SimpleQueueType = 0
	Transient SimpleQueueType = 1
)

// type ExchangeKind string

// const (
// 	Direct  ExchangeKind = "direct"
// 	Topic   ExchangeKind = "topic"
// 	Headers ExchangeKind = "headers"
// 	Fanout  ExchangeKind = "fanout"
// )

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	// kind ExchangeKind,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}
	isTransient := simpleQueueType == Transient
	queue, err := channel.QueueDeclare(queueName, !isTransient, isTransient, isTransient, false, nil)
	if err != nil {
		log.Fatalf("Could not declare queue: %v", err)
	}

	// Exchange and Queue types (durable, transient) should be independent — revisit ensuring exchanges are declared later
	// err = channel.ExchangeDeclare(exchange, string(kind), !isTransient, isTransient, false, false, nil)
	// if err != nil {
	// 	log.Fatalf("Could not declare exchange: %v", err)
	// }

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		log.Fatalf("Could not bind queue: %v", err)
	}

	return channel, queue, nil
}
