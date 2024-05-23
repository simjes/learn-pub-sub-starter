package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Need a username: %v", err)
	}

	pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)

	gameState := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(connection, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient, handlerPause(gameState))

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
			_, err := gameState.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
			}
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
			fmt.Println("Spamming not allowed yet")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
