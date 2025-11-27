package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		log.Printf("error Marshalling value: %v", val)
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes})
	if err != nil {
		log.Printf("error publishing message to the exchange")
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(val); err != nil {
		log.Printf("error encoding to gob: %v", val)
		return err
	}

	err := ch.PublishWithContext(context.Background(), exchange, key, false, false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        buffer.Bytes()})
	if err != nil {
		log.Printf("error publishing message to the exchange")
		return err
	}

	return nil
}