package handler

import (
	"log"

	"github.com/kwoodhouse93/trail-progress-worker/store"
)

type Handler struct {
	channel chan struct{}
}

func New() *Handler {
	return &Handler{
		channel: make(chan struct{}),
	}
}

func (h Handler) Func() store.NotificationHandler {
	return func(payload string) {
		log.Printf("handler: notification received - payload %q\n", payload)
		h.channel <- struct{}{}
	}
}

func (h Handler) Received() chan struct{} {
	return h.channel
}
