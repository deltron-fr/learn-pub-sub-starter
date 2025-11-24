package pubsub

import (
	"encoding/json"
	"log"

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
		return nil, amqp.Queue{}, err
	}

	err = subCh.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return subCh, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	subCh, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	ch, err := subCh.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for val := range ch {
			var t T
			err := json.Unmarshal(val.Body, &t)
			if err != nil {
				log.Printf("error unmarshalling message")
				return
			}
			handler(t)
			val.Ack(false)
		}
	}()

	return nil
}


