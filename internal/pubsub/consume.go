package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)



func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	subCh, err := conn.Channel()
	if err != nil {
		fmt.Println("Error creating channel")
		return nil, amqp.Queue{}, err
	}

	durableBool := false
	autoDelete := false
	exclusive := false

	switch queueType {
	case Durable:
		durableBool = true
	case Transient:
		autoDelete = true
		exclusive = true
	}

	queue, err := subCh.QueueDeclare(queueName, durableBool, autoDelete, exclusive, false, nil)
	if err != nil {
		fmt.Println("Error declaring queue")
		return nil, amqp.Queue{}, err
	}

	err = subCh.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Println("Error binding queue")
		return nil, amqp.Queue{}, err
	}

	return subCh, queue, nil
}