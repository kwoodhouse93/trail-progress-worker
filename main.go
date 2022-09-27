package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/kwoodhouse93/trail-progress-worker/handler"
	"github.com/kwoodhouse93/trail-progress-worker/processor"
	"github.com/kwoodhouse93/trail-progress-worker/store"
)

type Config struct {
	PostgresConnectionURL string        `required:"true" envconfig:"POSTGRES_CONNECTION_URL"`
	PostgresListenChannel string        `required:"true" envconfig:"POSTGRES_LISTEN_CHANNEL"`
	ProcessInterval       time.Duration `required:"true" envconfig:"PROCESS_INTERVAL"`
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
	defer store.Cleanup()

	handler := handler.New()
	processor := processor.New(store, handler.Received(), config.ProcessInterval)

	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		log.Println("sigint received")
		cancel()
	}()

	log.Println("starting store listener")
	go func() {
		err = store.Listen(ctx, config.PostgresListenChannel, handler.Func())
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("starting processor")
	err = processor.Serve(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal("shutting down")
}
