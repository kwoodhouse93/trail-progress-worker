package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/kelseyhightower/envconfig"
	"github.com/kwoodhouse93/trail-progress-worker/handler"
	"github.com/kwoodhouse93/trail-progress-worker/store"
)

type Config struct {
	PostgresConnectionURL string `required:"true" envconfig:"POSTGRES_CONNECTION_URL"`
	PostgresListenChannel string `required:"true" envconfig:"POSTGRES_LISTEN_CHANNEL"`
}

func main() {
	config := Config{}
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	store, err := store.New(config.PostgresConnectionURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := store.Cleanup()
		if err != nil {
			log.Println("error cleaning up store:", err)
		}
	}()

	handler := handler.New(store)

	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		cancel()
	}()

	log.Println("starting store listener")
	err = store.Listen(ctx, config.PostgresListenChannel, handler)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal("shutting down due to sigint")
}
