package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	messageBrokerConnectionString := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(messageBrokerConnectionString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not open channel %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Need a username: %v", err)
	}

	// pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)

	gameState := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient, handlerPause(gameState))
	pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, routing.ArmyMovesPrefix+".*", pubsub.Transient, handlerMove(channel, gameState))
	pubsub.SubscribeJSON(connection, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.Durable, handlerWar(channel, gameState))

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}

		if userInput[0] == "spawn" {
			err := gameState.CommandSpawn(userInput)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		if userInput[0] == "move" {
			mv, err := gameState.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
			}

			pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, mv)
			fmt.Printf("Published move: %v", mv)
			continue
		}

		if userInput[0] == "status" {
			gameState.CommandStatus()
			continue
		}

		if userInput[0] == "help" {
			gamelogic.PrintClientHelp()
			continue
		}

		if userInput[0] == "spam" {
			times, err := strconv.Atoi(userInput[1])
			if err != nil || times <= 0 {
				fmt.Println("Input was not a valid number")
				continue
			}

			for range times {
				logMsg := gamelogic.GetMaliciousLog()
				pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, routing.GameLog{
					CurrentTime: time.Now(),
					Username:    username,
					Message:     logMsg,
				})
			}

			continue
		}

		if userInput[0] == "quit" {
			gamelogic.PrintQuit()
			break
		}

		fmt.Println("I don't understand that command")
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done // Will block here until user hits ctrl+c

	fmt.Println("Shutting down")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(ch *amqp.Channel, gs *gamelogic.GameState) func(move gamelogic.ArmyMove) pubsub.AckType {
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)

		if outcome == gamelogic.MoveOutcomeMakeWar {
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix+"."+gs.Player.Username, gamelogic.RecognitionOfWar{
				Attacker: mv.Player,
				Defender: gs.Player,
			})

			if err != nil {
				return pubsub.NackRequeue
			}

			return pubsub.Ack
		}

		if outcome == gamelogic.MoveOutComeSafe {
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(ch *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)

		var message string

		if outcome == gamelogic.WarOutcomeOpponentWon || outcome == gamelogic.WarOutcomeYouWon {
			message = fmt.Sprintf("%s won a war against %s", winner, loser)
		}

		if outcome == gamelogic.WarOutcomeDraw {
			message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
		}

		err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+rw.Attacker.Username, routing.GameLog{
			CurrentTime: time.Now(),
			Username:    rw.Attacker.Username,
			Message:     message,
		})

		if err != nil {
			return pubsub.NackRequeue
		}

		if outcome == gamelogic.WarOutcomeNotInvolved {
			return pubsub.NackRequeue
		}

		if outcome == gamelogic.WarOutcomeNoUnits {
			return pubsub.NackDiscard
		}

		if outcome == gamelogic.WarOutcomeOpponentWon {
			return pubsub.Ack
		}

		if outcome == gamelogic.WarOutcomeYouWon {
			return pubsub.Ack
		}

		if outcome == gamelogic.WarOutcomeDraw {
			return pubsub.Ack
		}

		fmt.Printf("Error handling outcome, %v", outcome)
		return pubsub.NackDiscard
	}
}
