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

	fmt.Println("Peril game client connected to RabbitMQ!")
	rmqCh, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
		return
	}

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Printf("error getting username: %v", err)
		return
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient)

	if err != nil {
		log.Printf("error creating queue: %v", err)
		return
	}

	gameState := gamelogic.NewGameState(userName)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queue.Name,
		queueName,
		pubsub.Transient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Printf("error subscribing to pause: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"army_moves."+userName,
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gameState, rmqCh),
	)
	if err != nil {
		log.Printf("error subscribing to moves: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		"war.*",
		pubsub.Durable,
		handlerWar(gameState, rmqCh),
	)
	if err != nil {
		log.Printf("error subscribing to wars: %v", err)
		return
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				log.Printf("%v", err)
				return
			}
		case "move":
			move, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("%v", err)
				return
			}
			err = pubsub.PublishJSON(
				rmqCh,
				routing.ExchangePerilTopic,
				"army_moves."+userName,
				move,
			)
			if err != nil {
				log.Printf("error publishing move message: %v", err)
				return
			}
			fmt.Println("move published successfully")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		default:
			fmt.Println("Unknown command!")
		}
	}

}
