package handler

import (
	"context"
	"log"

	"github.com/kwoodhouse93/trail-progress-worker/store"
)

type Store interface {
	Process(ctx context.Context) error
}

func New(store Store) store.NotificationHandler {
	return func(payload string) {
		log.Printf("notification received - payload %q\n", payload)
		err := store.Process(context.Background())
		if err != nil {
			log.Fatalln("error processing store:", err)
		}
	}
}
