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
	"github.com/kwoodhouse93/trail-progress-worker/webhooks"
)

type PostgresConfig struct {
	ConnectionURL string `required:"true" envconfig:"CONNECTION_URL"`
	ListenChannel string `required:"true" envconfig:"LISTEN_CHANNEL"`
}

type StravaConfig struct {
	ClientID     int    `required:"true" envconfig:"CLIENT_ID"`
	ClientSecret string `required:"true" envconfig:"CLIENT_SECRET"`
	CallbackURL  string `required:"true" envconfig:"CALLBACK_URL"`
}

type Config struct {
	Postgres        PostgresConfig
	Strava          StravaConfig
	ProcessInterval time.Duration `required:"true" envconfig:"PROCESS_INTERVAL"`
}

func main() {
	config := Config{}
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	store, err := store.New(config.Postgres.ConnectionURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Cleanup()

	handler := handler.New()
	processor := processor.New(store, handler.Received(), config.ProcessInterval)

	log.Println("starting webhook subscription")
	subscription, err := webhooks.NewSubscription(config.Strava.ClientID, config.Strava.ClientSecret, config.Strava.CallbackURL, store)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		log.Println("sigint received")
		subscription.Close(context.Background())
		cancel()
	}()

	log.Println("starting store listener")
	go func() {
		err = store.Listen(ctx, config.Postgres.ListenChannel, handler.Func())
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
