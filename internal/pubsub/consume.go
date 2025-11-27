package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int
type Acktype int

const (
	Durable SimpleQueueType = iota
	Transient
)

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
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

	queue, err := subCh.QueueDeclare(
		queueName,
		durableBool,
		autoDelete,
		exclusive,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
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
	handler func(T) Acktype,
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
			ackType := handler(t)
			switch ackType {
			case Ack:
				val.Ack(false)
			case NackRequeue:
				val.Nack(false, true)
			case NackDiscard:
				val.Nack(false, false)
			}

		}
	}()

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
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
			buffer := bytes.NewBuffer(val.Body)
			decoder := gob.NewDecoder(buffer)

			if err := decoder.Decode(&t); err != nil {
				log.Printf("error unmarshalling message")
				return
			}

			ackType := handler(t)
			switch ackType {
			case Ack:
				val.Ack(false)
			case NackRequeue:
				val.Nack(false, true)
			case NackDiscard:
				val.Nack(false, false)
			}

		}
	}()

	return nil
}
