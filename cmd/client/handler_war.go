package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)



func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			message := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubGameLog(ch, rw, message)
			if err != nil {
				log.Printf("error publishing game log message: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			message := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := pubGameLog(ch, rw, message)
			if err != nil {
				log.Printf("error publishing game log message: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			message := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := pubGameLog(ch, rw, message)
			if err != nil {
				log.Printf("error publishing game log message: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("outcome unknown")
			return pubsub.NackDiscard
		}
	}
}

func pubGameLog(ch *amqp.Channel, rw gamelogic.RecognitionOfWar, msg string) error {
	gameLog := routing.GameLog{
		CurrentTime: time.Now(),
		Message: msg,
		Username: rw.Attacker.Username,
	}

	err := pubsub.PublishGob(
		ch,
		string(routing.ExchangePerilTopic),
		string(routing.GameLogSlug)+"."+rw.Attacker.Username,
		gameLog,
	)
	if err != nil {
		return err
	}
	return nil
}