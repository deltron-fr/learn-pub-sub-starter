package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.Durable)
	if err != nil {
		log.Printf("could not create bind topic exchange: %v", err)
		return
	}

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message!")

			err = pubsub.PublishJSON(
				rmqCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)

			if err != nil {
				log.Printf("could not publish pause message: %v", err)
			}
		case "resume":
			fmt.Println("Sending resume message!")

			err = pubsub.PublishJSON(
				rmqCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)

			if err != nil {
				log.Printf("could not publish resume message: %v", err)
				return
			}
		case "quit":
			fmt.Println("Shutting down...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command")
			continue
		}
	}
}
