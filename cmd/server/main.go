package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	fmt.Println("Peril game server connected to RabbitMQ!")

	rmqCh, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = pubsub.PublishJSON(
		rmqCh, 
		routing.ExchangePerilDirect, 
		routing.PauseKey, 
		routing.PlayingState{
			IsPaused: true,
			},
		)

	if err != nil {
		log.Printf("could not publish time: %v", err)
	}

	fmt.Println("Pause message sent!")

}
