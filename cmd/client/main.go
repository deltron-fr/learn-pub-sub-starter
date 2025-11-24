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

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Printf("error getting username: %v", err)
		return
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)

	_, _, err = pubsub.DeclareAndBind(
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

	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "spawn":
			err = gameState.CommandSpawn(words)
			if err != nil {
				log.Printf("%v", err)
			}
		case "move":
			_, err := gameState.CommandMove(words)
			if err != nil {
				log.Printf("%v", err)
			}
			fmt.Println("unit has been moved successfully")
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
